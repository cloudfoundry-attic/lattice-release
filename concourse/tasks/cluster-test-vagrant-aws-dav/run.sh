#!/bin/bash

set -ex

SYSTEM_DOMAIN=`cat deploy-vagrant-aws/system_domain`

tar zxf ltc-tar-build/ltc-*.tgz ltc-linux-amd64
./ltc-linux-amd64 target "$SYSTEM_DOMAIN"
./ltc-linux-amd64 test -v -t 5m || ./ltc-linux-amd64 test -v -t 10m
