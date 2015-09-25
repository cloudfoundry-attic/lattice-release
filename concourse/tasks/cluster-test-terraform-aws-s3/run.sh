#!/bin/bash

set -ex

DOT_LATTICE_DIR=$HOME/.lattice
LTC_VERSION=v$(cat ltc-tar-build/version)
TERRAFORM_TMP_DIR=$PWD/deploy-terraform-aws/terraform-tmp

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
    "active_blob_store": 1,
    "s3_blob_store": {
        "region": "${AWS_REGION}",
        "access_key": "${AWS_ACCESS_KEY_ID}",
        "secret_key": "${AWS_SECRET_ACCESS_KEY}",
        "bucket_name": "${S3_BUCKET_NAME}"
    }
}
EOF
popd

tar xzf ltc-tar-build/ltc-${LTC_VERSION}.tgz
./ltc-linux-amd64 test -v --timeout=5m
