#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice

mkdir -p $PWD/go/src/github.com/cloudfoundry-incubator $PWD/go/bin
ln -sf $LATTICE_DIR $PWD/go/src/github.com/cloudfoundry-incubator/lattice

export GOBIN=$PWD/go/bin
export GOPATH=$LATTICE_DIR/ltc/Godeps/_workspace:$PWD/go
export PATH=$GOBIN:$PATH

go install github.com/onsi/ginkgo/ginkgo
ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --trace --race $LATTICE_DIR/ltc

