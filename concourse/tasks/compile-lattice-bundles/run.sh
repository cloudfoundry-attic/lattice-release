#!/bin/bash

set -x -e

export LATTICE_VERSION=v$(cat ltc-tar-build/version)

LINUX_DIR=lattice-bundle-${LATTICE_VERSION}-linux
OSX_DIR=lattice-bundle-${LATTICE_VERSION}-osx

mkdir -p {$LINUX_DIR,$OSX_DIR}/vagrant

cp ltc-tar-build/ltc-linux-amd64 $LINUX_DIR/ltc
cp ltc-tar-build/ltc-darwin-amd64 $OSX_DIR/ltc

cp -r generate-terraform-templates/templates $LINUX_DIR/terraform
cp -r generate-terraform-templates/templates $OSX_DIR/terraform

cp generate-vagrantfile/Vagrantfile $LINUX_DIR/vagrant/
cp generate-vagrantfile/Vagrantfile $OSX_DIR/vagrant/

zip -r $LINUX_DIR $LINUX_DIR
zip -r $OSX_DIR $OSX_DIR
