#!/bin/bash -exu

function get_stack_output() {
  echo "$STACK_INFO" | jq -r "[ .Stacks[0].Outputs[] | { (.OutputKey): .OutputValue } | .$1 ] | add"
}

bosh -n target $BOSH_TARGET
bosh login admin $BOSH_PASSWORD

export BOSH_DIRECTOR_UUID=`bosh status --uuid`
export MANIFEST_FILE=$PWD/lattice/concourse/tasks/deploy-concourse/manifest.yml
export STACK_INFO=`aws cloudformation describe-stacks --stack-name "$CLOUDFORMATION_STACK_NAME"`

export SECURITY_GROUP_ID=$(get_stack_output "SecurityGroupID")
export SECURITY_GROUP_NAME=$(aws ec2 describe-security-groups --group-ids=$SECURITY_GROUP_ID | jq -r .SecurityGroups[0].GroupName)
export PRIVATE_SUBNET_ID=$(get_stack_output "PrivateSubnetID")
export ELB_NAME=$(get_stack_output "ElasticLoadBalancer")

REPLACED_MANIFEST_FILE=`interpolate $MANIFEST_FILE`
echo "$REPLACED_MANIFEST_FILE" > $MANIFEST_FILE

cat $MANIFEST_FILE
bosh deployment $MANIFEST_FILE
bosh upload stemcell https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent --skip-if-exists
bosh upload release "https://bosh.io/d/github.com/concourse/concourse?v=$CONCOURSE_VERSION" --skip-if-exists
bosh upload release "https://bosh.io/d/github.com/cloudfoundry-incubator/garden-linux-release?v=$GARDEN_LINUX_VERSION" --skip-if-exists
bosh -n deploy
