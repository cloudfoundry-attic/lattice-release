#!/bin/bash

set -ex

LATTICE_DIR=$PWD/lattice

LATTICE_VERSION=$(git -C $LATTICE_DIR describe --tags --always)
DIEGO_VERSION=$(cat $LATTICE_DIR/DIEGO_VERSION)
GARDEN_LINUX_VERSION=$(cat $LATTICE_DIR/GARDEN_LINUX_VERSION)
CF_VERSION=$(cat $LATTICE_DIR/CF_VERSION)
ROUTING_VERSION=$(cat $LATTICE_DIR/ROUTING_VERSION)

if [ "$RELEASE" = true ]; then
  LATTICE_VERSION=$(cat $LATTICE_DIR/Version)
fi

pushd $LATTICE_DIR/build/diego-release
	git checkout $DIEGO_VERSION
	git clean -xffd
	./scripts/update
popd

pushd $LATTICE_DIR/build/garden-linux-release
	git checkout $GARDEN_LINUX_VERSION
	git clean -xffd
  git submodule update --init --recursive
popd

pushd $LATTICE_DIR/build/cf-release
	git checkout $CF_VERSION
	git clean -xffd
	./update
popd

pushd $LATTICE_DIR/build/cf-routing-release
  git checkout $ROUTING_VERSION
  git clean -xffd
  ./scripts/update
popd

pushd $LATTICE_DIR/build/diego-release/src/github.com/cloudfoundry-incubator
  ln -sfn ../../../../.. lattice
popd

$LATTICE_DIR/cluster/scripts/compile $LATTICE_DIR/build/lattice-${LATTICE_VERSION}.tgz $LATTICE_VERSION

