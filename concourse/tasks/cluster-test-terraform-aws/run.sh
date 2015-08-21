#!/bin/bash

set -x -e

export DOT_LATTICE_DIR=$HOME/.lattice
export LATTICE_VERSION=$(cat lattice-tar-experimental/version)
export LTC_VERSION=$(cat ltc-tar-experimental/version)
export TERRAFORM_TMP_DIR=$PWD/terraform-tmp

mkdir -p $TERRAFORM_TMP_DIR $DOT_LATTICE_DIR

cat <<< "$AWS_SSH_PRIVATE_KEY" > $TERRAFORM_TMP_DIR/key.pem

cat << EOF > $TERRAFORM_TMP_DIR/lattice.tf
{
    "module": {
        "lattice-aws": {
            "source": "../lattice/terraform/aws",
            "lattice_tar_source": "../lattice-tar-experimental/lattice-v${LATTICE_VERSION}.tgz",
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
