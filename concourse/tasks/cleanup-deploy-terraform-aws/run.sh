#!/bin/bash

set -ex

pushd $PWD/deploy-terraform-aws/terraform-tmp
    terraform get -update
    terraform destroy -force || terraform destroy -force
popd
