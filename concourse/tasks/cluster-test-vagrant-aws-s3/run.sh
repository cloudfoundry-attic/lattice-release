#!/bin/bash

set -ex

SYSTEM_DOMAIN=`cat deploy-vagrant-aws/system_domain`

mkdir -p ~/.lattice

cat > ~/.lattice/config.json <<EOF
{
  "target": "${SYSTEM_DOMAIN}",
  "active_blob_store": 1,
  "s3_blob_store": {
    "region": "${AWS_REGION}",
    "access_key": "${AWS_ACCESS_KEY_ID}",
    "secret_key": "${AWS_SECRET_ACCESS_KEY}",
    "bucket_name": "${S3_BUCKET_NAME}"
  }
}
EOF

tar zxf ltc-tar-build/ltc-*.tgz ltc-linux-amd64
./ltc-linux-amd64 test -v -t 5m || ./ltc-linux-amd64 test -v -t 10m
