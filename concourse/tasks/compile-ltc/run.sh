#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice

mkdir -p $PWD/go/src/github.com/cloudfoundry-incubator $LATTICE_DIR/build
ln -sf $LATTICE_DIR $PWD/go/src/github.com/cloudfoundry-incubator/lattice

export LATTICE_VERSION=$(git -C $LATTICE_DIR describe)
export DIEGO_VERSION=$(cat $LATTICE_DIR/DIEGO_VERSION)

export GOPATH=$LATTICE_DIR/ltc/Godeps/_workspace:$PWD/go

GOARCH=amd64 GOOS=linux go build \
    -ldflags \
        "-X github.com/cloudfoundry-incubator/lattice/ltc/setup_cli.latticeVersion $LATTICE_VERSION
         -X github.com/cloudfoundry-incubator/lattice/ltc/setup_cli.diegoVersion $DIEGO_VERSION" \
    -o ltc-linux-amd64 \
    github.com/cloudfoundry-incubator/lattice/ltc

GOARCH=amd64 GOOS=darwin go build \
    -ldflags \
        "-X github.com/cloudfoundry-incubator/lattice/ltc/setup_cli.latticeVersion $LATTICE_VERSION
         -X github.com/cloudfoundry-incubator/lattice/ltc/setup_cli.diegoVersion $DIEGO_VERSION" \
    -o ltc-darwin-amd64 \
    github.com/cloudfoundry-incubator/lattice/ltc

tar czf $LATTICE_DIR/build/ltc-${LATTICE_VERSION}.tgz ltc-linux-amd64 ltc-darwin-amd64
