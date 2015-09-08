#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice
export LATTICE_VERSION=$(git -C $LATTICE_DIR describe --tags --always)
if [ "$RELEASE" = true ]; then
  export LATTICE_VERSION=$(cat $LATTICE_DIR/Version)
fi

export LATTICE_URL="${LATTICE_URL_BASE}/lattice-${LATTICE_VERSION}.tgz"

( echo "LATTICE_URL = '${LATTICE_URL}'"; cat $LATTICE_DIR/Vagrantfile ) > Vagrantfile
