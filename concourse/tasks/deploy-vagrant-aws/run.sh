#!/bin/bash

set -ex

LATTICE_DIR=$PWD/lattice

export AWS_SSH_PRIVATE_KEY_PATH=$PWD/vagrant.pem
export AWS_INSTANCE_NAME=concourse-vagrant

cat <<< "$AWS_SSH_PRIVATE_KEY" > "$AWS_SSH_PRIVATE_KEY_PATH"

curl -LO https://dl.bintray.com/mitchellh/vagrant/vagrant_1.7.4_x86_64.deb
dpkg -i vagrant_1.7.4_x86_64.deb

while ! vagrant plugin install vagrant-aws; do
  sleep 5
done

vagrant box add lattice/ubuntu-trusty-64 --provider=aws

cp lattice-tar-build/lattice-*.tgz $LATTICE_DIR/lattice.tgz
pushd $LATTICE_DIR
  vagrant up --provider=aws
  export $(vagrant ssh -c "grep SYSTEM_DOMAIN /var/lattice/setup/lattice-environment" | egrep -o '(SYSTEM_DOMAIN=.+\.io)')
popd

sleep 120

echo $SYSTEM_DOMAIN > system_domain
