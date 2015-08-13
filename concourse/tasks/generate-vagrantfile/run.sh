#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice
export LATTICE_VERSION=$(git -C $LATTICE_DIR describe)

( echo "LATTICE_URL = '$LATTICE_VERSION'"; cat $LATTICE_DIR/Vagrantfile ) > Vagrantfile-$LATTICE_VERSION
