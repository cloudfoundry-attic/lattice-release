#!/bin/bash

set -ex

LATTICE_VERSION=v$(cat ltc-tar-build/version)

LINUX_DIR=lattice-bundle-${LATTICE_VERSION}-linux
OSX_DIR=lattice-bundle-${LATTICE_VERSION}-osx

mkdir -p {$LINUX_DIR,$OSX_DIR}/vagrant

( cd ltc-tar-build && tar xvzf ltc-${LATTICE_VERSION}.tgz )
cp ltc-tar-build/ltc-linux-amd64 $LINUX_DIR/ltc
cp ltc-tar-build/ltc-darwin-amd64 $OSX_DIR/ltc

cp -r generate-terraform-templates/templates $LINUX_DIR/terraform
cp -r generate-terraform-templates/templates $OSX_DIR/terraform

cp generate-vagrantfile/Vagrantfile $LINUX_DIR/vagrant/
cp generate-vagrantfile/Vagrantfile $OSX_DIR/vagrant/

zip -r ${LINUX_DIR}.zip $LINUX_DIR
zip -r ${OSX_DIR}.zip $OSX_DIR

git -C lattice rev-parse HEAD > commit
