#!/bin/bash -exu

function setup() {
  bosh-init --version
  aws --version
  aws configure list
  export DIR=$(cd $(dirname $0) && pwd)
  export STACK_INFO=`aws cloudformation describe-stacks --stack-name "$CLOUDFORMATION_STACK_NAME"`

  export PRIVATE_KEY_PATH=/tmp/private-key
  echo "$PRIVATE_KEY" > $PRIVATE_KEY_PATH
  chmod 0600 $PRIVATE_KEY_PATH
}

function get_stack_output() {
  echo "$STACK_INFO" | jq -r "[ .Stacks[0].Outputs[] | { (.OutputKey): .OutputValue } | .$1 ] | add"
}

function build_manifest() {
  export BOSH_DIRECTOR_RELEASE_URL=https://bosh.io/d/github.com/cloudfoundry/bosh
  export BOSH_CPI_RELEASE_URL=https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release
  export BOSH_STEMCELL_URL=https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent

  export ENVIRONMENT_NAME=$CLOUDFORMATION_STACK_NAME
  export VPC_ID=$(get_stack_output "VPCID")
  export SECURITY_GROUP_ID=$(get_stack_output "SecurityGroupID")
  export SECURITY_GROUP_NAME=$(aws ec2 describe-security-groups --group-ids=$SECURITY_GROUP_ID | jq -r .SecurityGroups[0].GroupName)
  export PRIVATE_SUBNET_ID=$(get_stack_output "PublicSubnetID")
  export ELASTIC_IP=$(get_stack_output "ElasticIP")
  export AVAILABILITY_ZONE=$(get_stack_output "AvailabilityZone")

  export BOSH_DIRECTOR_RELEASE_SHA=$(curl -L $BOSH_DIRECTOR_RELEASE_URL 2>/dev/null | shasum -b -a 1 | cut -d ' ' -f1)
  export BOSH_CPI_RELEASE_SHA=$(curl -L $BOSH_CPI_RELEASE_URL 2>/dev/null | shasum -b -a 1 | cut -d ' ' -f1)
  export BOSH_STEMCELL_SHA=$(curl -L $BOSH_STEMCELL_URL 2>/dev/null | shasum -b -a 1 | cut -d ' ' -f1)

  env | sort | grep =

  interpolate $DIR/bosh.yml > /tmp/bosh.yml

  cat /tmp/bosh.yml
}

function deploy() {
  echo "will deploy using $KEY_NAME"
  bosh-init deploy /tmp/bosh.yml
  cat /tmp/bosh-state.json
}

function update_password() {
  bosh -n target "$ELASTIC_IP"
  bosh -n login admin admin
  bosh create user admin "$BOSH_PASSWORD"
}

setup
build_manifest
deploy
update_password
