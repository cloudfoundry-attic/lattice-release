#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice
export LATTICE_VERSION=$(cat lattice/Version)

export LATTICE_TAR_PATH=$LATTICE_DIR/build/lattice.tgz
export TF_WORKING_DIR=$HOME/terraform-work # ephemeral folder in docker context
mkdir -p $TF_WORKING_DIR

trap cleanup EXIT

cp concourse-key/concourse-test-v0.pem $TF_WORKING_DIR/

printf '{
    "module":{
        "lattice-aws":{
            "source":"%s",
            "local_lattice_tar_path": "%s",
            "num_cells": "1",
            "aws_access_key": "%s",
            "aws_secret_key": "%s",
            "aws_region": "us-east-1",
            "aws_key_name": "concourse-test",
            "aws_ssh_private_key_file": "%s/concourse-test-v0.pem"
        }
    }
    "output": {
        "lattice_target": {
            "value": "${module.lattice-aws.lattice_target}"
        },
        "lattice_username": {
            "value": "${module.lattice-aws.lattice_username}"
        },
        "lattice_password": {
            "value": "${module.lattice-aws.lattice_password}"
        }
    }
}' \
"$LATTICE_DIR/terraform/aws" \
"$LATTICE_TAR_PATH" \
"$AWS_ACCESS_KEY_ID" "$AWS_SECRET_ACCESS_KEY" "$TF_WORKING_DIR" \
"$terraform_outputs" \
> $TF_WORKING_DIR/lattice.tf

echo "== lattice.tf =="
    cat $TF_WORKING_DIR/lattice.tf
echo "===="


pushd $TF_WORKING_DIR
    terraform get -update
    terraform apply || { echo "=====>First terraform apply failed. Retrying..."; terraform apply; }
popd

echo "Sleeping for 3 minutes.."
sleep 180

tar xzf ltc-tar-experimental
LTC_PATH = $PWD/ltc-darwin-amd64
mkdir -p .lattice

echo "=========================Lattice Integration Tests=============================\n"

printf "{\"target\":\"%s\",\"username\":\"%s\",\"password\":\"%s\",\"dav_blob_store\":{\"host\":\"%s\",\"port\":\"%s\",\"username\":\"%s\",\"password\":\"%s\"}}" \
    "$(cd $TF_WORKING_DIR && terraform output lattice_target)" \
    "$(cd $TF_WORKING_DIR && terraform output lattice_username)" \
    "$(cd $TF_WORKING_DIR && terraform output lattice_password)" \
    "$(cd $TF_WORKING_DIR && terraform output lattice_target)" \
    "8444" \
    "$(cd $TF_WORKING_DIR && terraform output lattice_username)" \
    "$(cd $TF_WORKING_DIR && terraform output lattice_password)" | 
    json_pp > $PWD/.lattice/config.json

ltc test -v --timeout=5m

echo "===============================================================================\n"
