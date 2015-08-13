#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice
export AWS_SSH_PRIVATE_KEY_PATH=$PWD/concourse-key/concourse-test-v0.pem

mkdir vagrant-up-tmp
cd vagrant-up-tmp

curl -LO https://dl.bintray.com/mitchellh/vagrant/vagrant_1.7.4_x86_64.deb && dpkg -i vagrant_1.7.4_x86_64.deb

vagrant plugin install vagrant-aws
vagrant box add lattice/ubuntu-trusty-64 --provider=aws

(cp ../lattice-tar-experimental/lattice-*.tgz lattice.tgz && cp ../lattice/Vagrantfile ./ && vagrant up --provider=aws)
export $(vagrant ssh -c "grep SYSTEM_DOMAIN /var/lattice/setup/lattice-environment" | egrep -o '(SYSTEM_DOMAIN=.+\.io)')

cd ..

tar zxf ltc-tar-experimental/ltc-*.tgz
./ltc-linux-amd64 target $SYSTEM_DOMAIN
$PWD/ltc-linux-amd64 test -v -t 10m
