#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice
export DOT_LATTICE_DIR=$HOME/.lattice
export LATTICE_VERSION=$(cat lattice-tar-experimental/version)
export LTC_VERSION=$(cat ltc-tar-experimental/version)
export TERRAFORM_TMP_DIR=$PWD/terraform-tmp

mkdir -p $TERRAFORM_TMP_DIR $DOT_LATTICE_DIR

cat <<< "$GCE_SSH_PRIVATE_KEY" > $TERRAFORM_TMP_DIR/concourse-test.pem
cat <<< "$GCE_ACCOUNT_FILE_JSON" > $TERRAFORM_TMP_DIR/gce-account.json

cat << EOF > $TERRAFORM_TMP_DIR/lattice.tf
{
    "module": {
        "lattice-google": {
            "source": "${LATTICE_DIR}/terraform/google",
            "lattice_username": "user",
            "lattice_password": "pass",
            "local_lattice_tar_path": "${PWD}/lattice-tar-experimental/lattice-v${LATTICE_VERSION}.tgz",
            "gce_ssh_user": "pivotal",
            "gce_ssh_private_key_file": "${TERRAFORM_TMP_DIR}/concourse-test.pem",
            "gce_project": "${GCE_PROJECT_NAME}",
            "gce_account_file": "${TERRAFORM_TMP_DIR}/gce-account.json"
        }
    },
    "output": {
        "lattice_target": {
            "value": "\${module.lattice-google.lattice_target}"
        },
        "lattice_username": {
            "value": "\${module.lattice-google.lattice_username}"
        },
        "lattice_password": {
            "value": "\${module.lattice-google.lattice_password}"
        }
    }
}
EOF

cleanup() { ( cd $TERRAFORM_TMP_DIR && terraform destroy -force || terraform destroy -force ) }
trap cleanup EXIT

pushd $TERRAFORM_TMP_DIR
    terraform get -update
    terraform apply || terraform apply
popd

sleep 60

tar xzf ltc-tar-experimental/ltc-v${LTC_VERSION}.tgz

pushd $TERRAFORM_TMP_DIR
    LATTICE_TARGET=$(terraform output lattice_target)
    LATTICE_USERNAME=$(terraform output lattice_username)
    LATTICE_PASSWORD=$(terraform output lattice_password)
    cat << EOF > $DOT_LATTICE_DIR/config.json
{
    "target": "${LATTICE_TARGET}",
    "username": "${LATTICE_USERNAME}",
    "password": "${LATTICE_PASSWORD}",
    "dav_blob_store": {
        "host": "${LATTICE_TARGET}",
        "port": "8444",
        "username": "${LATTICE_USERNAME}",
        "password": "${LATTICE_PASSWORD}"
    }
}
EOF
popd

$PWD/ltc-linux-amd64 test -v --timeout=5m
