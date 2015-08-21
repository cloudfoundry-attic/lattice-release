#!/bin/bash

set -x -e

export LATTICE_DIR=$PWD/lattice
export LATTICE_VERSION=$(git -C $LATTICE_DIR describe)

for PROVIDER in aws digitalocean google openstack; do
	INPUT=${LATTICE_DIR}/terraform/${PROVIDER}/example/lattice.${PROVIDER}.tf
	OUTPUT=templates/${PROVIDER}/lattice-${LATTICE_VERSION}.${PROVIDER}.tf
 	mkdir -p `dirname $OUTPUT`

	SOURCE="github.com/cloudfoundry-incubator/lattice//terraform//${PROVIDER}?ref=${LATTICE_VERSION}"
	LATTICE_TAR_URL="https://s3.amazonaws.com/${S3_LATTICE_PATH}/backend/lattice-${LATTICE_VERSION}.tgz"

	(
		sed 's@# source =.*$@source = "'${SOURCE}'"@' | \
		sed 's@# lattice_tar_source =.*$@lattice_tar_source = "'${LATTICE_TAR_URL}'"@'
	) < $INPUT > $OUTPUT
done
