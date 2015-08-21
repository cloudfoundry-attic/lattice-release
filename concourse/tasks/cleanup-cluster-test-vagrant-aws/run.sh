#!/bin/bash

set -x -e

export AWS_SSH_PRIVATE_KEY_PATH=$PWD/cluster-test-vagrant-aws/vagrant.pem
export VAGRANT_TMP_DIR=$PWD/cluster-test-vagrant-aws/vagrant-tmp
cat <<< "$AWS_SSH_PRIVATE_KEY" > "$AWS_SSH_PRIVATE_KEY_PATH"

curl -LO https://dl.bintray.com/mitchellh/vagrant/vagrant_1.7.4_x86_64.deb
dpkg -i vagrant_1.7.4_x86_64.deb

while ! vagrant plugin install vagrant-aws; do
  sleep 5
done

vagrant box add lattice/ubuntu-trusty-64 --provider=aws

( cd $VAGRANT_TMP_DIR && vagrant destroy -f )
