#!/bin/bash

set -ex

LATTICE_DIR=$PWD/lattice
LATTICE_VERSION=$(git -C $LATTICE_DIR describe --tags --always)
DIEGO_VERSION=$(cat $LATTICE_DIR/DIEGO_VERSION)

if [ "$RELEASE" = true ]; then
  LATTICE_VERSION=$(cat $LATTICE_DIR/Version)
fi

mkdir -p $PWD/go/src/github.com/cloudfoundry-incubator $LATTICE_DIR/build
ln -sf $LATTICE_DIR $PWD/go/src/github.com/cloudfoundry-incubator/lattice

export GOPATH=$LATTICE_DIR/ltc/Godeps/_workspace:$PWD/go
export GOARCH=amd64

GOOS=linux go build \
    -ldflags "-X github.com/cloudfoundry-incubator/lattice/ltc/setup_cli.latticeVersion $LATTICE_VERSION
              -X github.com/cloudfoundry-incubator/lattice/ltc/setup_cli.diegoVersion $DIEGO_VERSION" \
    -o ltc-linux-amd64 \
    github.com/cloudfoundry-incubator/lattice/ltc

GOOS=darwin go build \
    -ldflags "-X github.com/cloudfoundry-incubator/lattice/ltc/setup_cli.latticeVersion $LATTICE_VERSION
              -X github.com/cloudfoundry-incubator/lattice/ltc/setup_cli.diegoVersion $DIEGO_VERSION" \
    -o ltc-darwin-amd64 \
    github.com/cloudfoundry-incubator/lattice/ltc

tar czf $LATTICE_DIR/build/ltc-${LATTICE_VERSION}.tgz ltc-linux-amd64 ltc-darwin-amd64
