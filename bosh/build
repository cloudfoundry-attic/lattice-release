#!/bin/bash

set -e

if [[ -z $1 ]] || [[ ! -d $(dirname "$1") ]]; then
  echo "Usage:"
  echo -e "\t$0 /new/or/existing/assets/directory"
  exit 1
fi

assets_dir=$(mkdir -p "$1" && cd "$1" && pwd)
lattice_release_dir=$(cd `dirname $0` && cd .. && pwd)

rm -rf $assets_dir/{releases,versions}
mkdir $assets_dir/{releases,versions}

git -C "$lattice_release_dir" rev-parse HEAD > $assets_dir/versions/LATTICE_RELEASE_IMAGE

for release in diego garden-linux cf-routing etcd cf-lattice; do
  pushd $lattice_release_dir/$release-release >/dev/null
    yes yes | bosh -n reset release
    bosh -n create release --name $release --version 0 --with-tarball --force
    mv dev_releases/$release/$release-0.tgz $assets_dir/releases/
    version_file="$(echo ${release//-/_} | tr [:lower:] [:upper:])_RELEASE"
    git describe --tags --always > $assets_dir/versions/$version_file
  popd >/dev/null
done
