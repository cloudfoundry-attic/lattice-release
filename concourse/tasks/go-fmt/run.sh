#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice

pushd $LATTICE_DIR/$GO_FMT_PATH
	test -z "$(go fmt ./...)"
popd



