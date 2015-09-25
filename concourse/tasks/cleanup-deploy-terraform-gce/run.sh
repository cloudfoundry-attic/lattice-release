#!/bin/bash

set -ex

pushd $PWD/deploy-terraform-gce/terraform-tmp
    terraform get -update
    terraform destroy -force || terraform destroy -force
popd

