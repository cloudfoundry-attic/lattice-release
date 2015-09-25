#!/bin/bash

set -ex

LATTICE_DIR=$PWD/lattice

pushd $LATTICE_DIR/$GO_FMT_PATH
	test -z "$(go fmt ./...)"
popd

