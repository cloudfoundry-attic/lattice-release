#!/bin/bash

set -ex

LATTICE_DIR=$PWD/deploy-vagrant-aws/lattice

export AWS_SSH_PRIVATE_KEY_PATH=$PWD/deploy-vagrant-aws/vagrant.pem
cat <<< "$AWS_SSH_PRIVATE_KEY" > "$AWS_SSH_PRIVATE_KEY_PATH"

curl -LO https://dl.bintray.com/mitchellh/vagrant/vagrant_1.7.4_x86_64.deb
dpkg -i vagrant_1.7.4_x86_64.deb

while ! vagrant plugin install vagrant-aws; do
  sleep 5
done

vagrant box add lattice/ubuntu-trusty-64 --provider=aws

( cd $LATTICE_DIR && vagrant destroy -f )
