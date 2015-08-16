#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice
export DIEGO_RELEASE_DIR=$PWD/lattice/build/diego-release
export GOPATH=$DIEGO_RELEASE_DIR
export PATH=$GOPATH/bin:$PATH

export LATTICE_VERSION=$(git -C $LATTICE_DIR describe)
export DIEGO_VERSION=$(cat $LATTICE_DIR/DIEGO_VERSION)
export CF_VERSION=$(cat $LATTICE_DIR/CF_VERSION)

pushd $DIEGO_RELEASE_DIR
	git checkout $DIEGO_VERSION
	git clean -xffd
	./scripts/update
popd

pushd $LATTICE_DIR/build/cf-release
	git checkout $CF_VERSION
	git clean -xffd
	./update
popd

$LATTICE_DIR/cluster/scripts/compile \
    $LATTICE_DIR/build/lattice-build \
    $LATTICE_DIR/build/diego-release \
    $LATTICE_DIR/build/cf-release \
    $LATTICE_DIR

echo $LATTICE_VERSION > $LATTICE_DIR/build/lattice-build/common/LATTICE_VERSION
echo $DIEGO_VERSION > $LATTICE_DIR/build/lattice-build/common/DIEGO_VERSION
echo $CF_VERSION > $LATTICE_DIR/build/lattice-build/common/CF_VERSION

tar czf $LATTICE_DIR/build/lattice.tgz -C $LATTICE_DIR/build lattice-build

mv $LATTICE_DIR/build/lattice.tgz $LATTICE_DIR/build/lattice-${LATTICE_VERSION}.tgz

