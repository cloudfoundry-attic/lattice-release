#!/bin/bash

set -x -e

export DOT_LATTICE_DIR=$HOME/.lattice
export LATTICE_VERSION=v$(cat lattice-tar-build/version)
export LTC_VERSION=v$(cat ltc-tar-build/version)
export TERRAFORM_TMP_DIR=$PWD/terraform-tmp

mkdir -p $TERRAFORM_TMP_DIR $DOT_LATTICE_DIR

cat <<< "$GCE_SSH_PRIVATE_KEY" > $TERRAFORM_TMP_DIR/key.pem
cat <<< "$GCE_ACCOUNT_FILE_JSON" > $TERRAFORM_TMP_DIR/gce-account.json

cat << EOF > $TERRAFORM_TMP_DIR/lattice.tf
{
    "module": {
        "lattice-google": {
            "source": "../lattice/terraform/google",
            "lattice_tar_source": "../lattice-tar-build/lattice-${LATTICE_VERSION}.tgz",
            "lattice_username": "user",
            "lattice_password": "pass",
            "gce_ssh_user": "pivotal",
            "gce_ssh_private_key_file": "./key.pem",
            "gce_project": "${GCE_PROJECT_NAME}",
            "gce_account_file": "./gce-account.json"
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

pushd $TERRAFORM_TMP_DIR
    terraform get -update
    terraform apply || terraform apply
popd

sleep 60

tar xzf ltc-tar-build/ltc-${LTC_VERSION}.tgz

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
