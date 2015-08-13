#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice

mkdir -p $PWD/go/src/github.com/cloudfoundry-incubator $PWD/go/bin
ln -sf $LATTICE_DIR $PWD/go/src/github.com/cloudfoundry-incubator/lattice

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH

go get github.com/onsi/ginkgo/ginkgo

pushd $PWD/go/src/github.com/cloudfoundry-incubator/lattice/cell-helpers/tee2metron
  go get -v -t ./...
  ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --trace --race .
popd
