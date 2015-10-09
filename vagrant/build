#!/bin/bash

set -e

lattice_release_dir=$(cd `dirname $0` && cd .. && pwd)
vagrant_dir=$lattice_release_dir/vagrant

$lattice_release_dir/bosh/setup $vagrant_dir/assets
$lattice_release_dir/bosh/build $vagrant_dir/assets

pushd "$vagrant_dir" >/dev/null
  packer build $@ lattice.json
popd >/dev/null
