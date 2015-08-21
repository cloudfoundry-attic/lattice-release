#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice
export LATTICE_VERSION=$(git -C $LATTICE_DIR describe)
export LATTICE_URL="https://s3.amazonaws.com/${S3_LATTICE_PATH}/backend/lattice-${LATTICE_VERSION}.tgz"

( echo "LATTICE_URL = '${LATTICE_URL}'"; cat $LATTICE_DIR/Vagrantfile ) > Vagrantfile
