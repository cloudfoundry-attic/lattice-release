#!/bin/bash

set -ex

LATTICE_VERSION=v$(cat lattice-tar-build/version)
TERRAFORM_TMP_DIR=$PWD/terraform-tmp

mkdir -p $TERRAFORM_TMP_DIR

cat <<< "$AWS_SSH_PRIVATE_KEY" > $TERRAFORM_TMP_DIR/key.pem

cat > $TERRAFORM_TMP_DIR/lattice.tf <<EOF
{
    "module": {
        "lattice-aws": {
            "source": "../lattice/terraform/aws",
            "lattice_tar_source": "../lattice-tar-build/lattice-${LATTICE_VERSION}.tgz",
            "lattice_username": "user",
            "lattice_password": "pass",
            "num_cells": "1",
            "aws_access_key": "${AWS_ACCESS_KEY_ID}",
            "aws_secret_key": "${AWS_SECRET_ACCESS_KEY}",
            "aws_region": "us-east-1",
            "aws_key_name": "${AWS_SSH_PRIVATE_KEY_NAME}",
            "aws_ssh_private_key_file": "./key.pem"
        }
    },
    "output": {
        "lattice_target": {
            "value": "\${module.lattice-aws.lattice_target}"
        },
        "lattice_username": {
            "value": "\${module.lattice-aws.lattice_username}"
        },
        "lattice_password": {
            "value": "\${module.lattice-aws.lattice_password}"
        }
    }
}
EOF

pushd $TERRAFORM_TMP_DIR
    terraform get -update
    terraform apply || terraform apply
popd

sleep 60
