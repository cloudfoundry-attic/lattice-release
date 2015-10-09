#!/bin/bash

set -e

lattice_release_dir=$(cd `dirname $0` && cd .. && pwd)
terraform_dir=$lattice_release_dir/terraform

$lattice_release_dir/bosh/setup $terraform_dir/assets
$lattice_release_dir/bosh/build $terraform_dir/assets

pushd "$terraform_dir" >/dev/null
  packer build $@ lattice.json
popd >/dev/null
