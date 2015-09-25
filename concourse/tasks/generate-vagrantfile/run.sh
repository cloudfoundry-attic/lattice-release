#!/bin/bash

set -ex

LATTICE_DIR=$PWD/lattice
LATTICE_VERSION=$(git -C $LATTICE_DIR describe --tags --always)

if [ "$RELEASE" = true ]; then
  LATTICE_VERSION=$(cat $LATTICE_DIR/Version)
fi

LATTICE_URL="${LATTICE_URL_BASE}/lattice-${LATTICE_VERSION}.tgz"

( echo "LATTICE_URL = '${LATTICE_URL}'"; cat $LATTICE_DIR/Vagrantfile ) > Vagrantfile
