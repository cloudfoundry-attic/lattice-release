#!/bin/bash

set -x -e

export TERRAFORM_TMP_DIR=$PWD/deploy-terraform-aws/terraform-tmp

pushd $TERRAFORM_TMP_DIR
    terraform get -update
    terraform destroy -force || terraform destroy -force
popd
