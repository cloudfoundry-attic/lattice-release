#!/bin/bash

set -ex

DOT_LATTICE_DIR=$HOME/.lattice
LTC_VERSION=v$(cat ltc-tar-build/version)
TERRAFORM_TMP_DIR=$PWD/deploy-terraform-gce/terraform-tmp

mkdir -p $DOT_LATTICE_DIR

pushd $TERRAFORM_TMP_DIR
    LATTICE_TARGET=$(terraform output lattice_target)
    LATTICE_USERNAME=$(terraform output lattice_username)
    LATTICE_PASSWORD=$(terraform output lattice_password)
    cat > $DOT_LATTICE_DIR/config.json <<EOF
{
    "target": "${LATTICE_TARGET}",
    "username": "${LATTICE_USERNAME}",
    "password": "${LATTICE_PASSWORD}",
    "active_blob_store": 0,
    "dav_blob_store": {
        "host": "${LATTICE_TARGET}",
        "port": "8444",
        "username": "${LATTICE_USERNAME}",
        "password": "${LATTICE_PASSWORD}"
    }
}
EOF
popd

tar xzf ltc-tar-build/ltc-${LTC_VERSION}.tgz
./ltc-linux-amd64 test -v --timeout=5m
