#!/bin/bash -exu

bosh -n target $BOSH_TARGET
bosh login $BOSH_USER $BOSH_PASSWORD

export BOSH_DIRECTOR_UUID=`bosh status --uuid`
export MANIFEST_FILE=$PWD/lattice/concourse/tasks/concourse-deploy/manifest.yml

REPLACED_MANIFEST_FILE=`interpolate $MANIFEST_FILE`
echo "$REPLACED_MANIFEST_FILE" > $MANIFEST_FILE

cat $MANIFEST_FILE
bosh deployment $MANIFEST_FILE
bosh upload stemcell https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent --skip-if-exists
bosh upload release "https://bosh.io/d/github.com/concourse/concourse?v=$CONCOURSE_VERSION" --skip-if-exists
bosh upload release "https://bosh.io/d/github.com/cloudfoundry-incubator/garden-linux-release?v=$GARDEN_LINUX_VERSION" --skip-if-exists
bosh -n deploy
