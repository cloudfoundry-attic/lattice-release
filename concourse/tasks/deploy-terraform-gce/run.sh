#!/bin/bash

set -ex

LATTICE_VERSION=v$(cat lattice-tar-build/version)
TERRAFORM_TMP_DIR=$PWD/terraform-tmp

mkdir -p $TERRAFORM_TMP_DIR

cat <<< "$GCE_SSH_PRIVATE_KEY" > $TERRAFORM_TMP_DIR/key.pem
cat <<< "$GCE_ACCOUNT_FILE_JSON" > $TERRAFORM_TMP_DIR/gce-account.json

apt-get install -y uuid
UUID=$(uuid)

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
            "gce_account_file": "./gce-account.json",
            "lattice_namespace": "concourse-${UUID}"
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
