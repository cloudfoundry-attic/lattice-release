#!/bin/bash

set -x -e

export TERRAFORM_TMP_DIR=$PWD/cluster-test-terraform-aws/terraform-tmp

pushd $TERRAFORM_TMP_DIR
    terraform get -update
    terraform destroy -force || terraform destroy -force
popd

