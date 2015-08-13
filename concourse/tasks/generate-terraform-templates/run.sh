#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice
export LATTICE_VERSION=$(git -C $LATTICE_DIR describe)

sed 's/source = "github\.com.*$/source = "github\.com\/cloudfoundry-incubator\/lattice\/\/terraform\/\/aws\?ref='"$LATTICE_VERSION"'"/' \
	< $LATTICE_DIR/terraform/aws/example/lattice.aws.tf > lattice-$LATTICE_VERSION.aws.tf
sed 's/source = "github\.com.*$/source = "github\.com\/cloudfoundry-incubator\/lattice\/\/terraform\/\/digitalocean\?ref='"$LATTICE_VERSION"'"/' \
	< $LATTICE_DIR/terraform/digitalocean/example/lattice.digitalocean.tf > lattice-$LATTICE_VERSION.digitalocean.tf
sed 's/source = "github\.com.*$/source = "github\.com\/cloudfoundry-incubator\/lattice\/\/terraform\/\/google\?ref='"$LATTICE_VERSION"'"/' \
	< $LATTICE_DIR/terraform/google/example/lattice.google.tf > lattice-$LATTICE_VERSION.google.tf
sed 's/source = "github\.com.*$/source = "github\.com\/cloudfoundry-incubator\/lattice\/\/terraform\/\/openstack\?ref='"$LATTICE_VERSION"'"/' \
	< $LATTICE_DIR/terraform/openstack/example/lattice.openstack.tf > lattice-$LATTICE_VERSION.openstack.tf
