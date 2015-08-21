#!/bin/bash -ux

function stack_exists() {
  aws cloudformation describe-stacks --stack-name $1 >/dev/null 2>/dev/null
  return $?
}

function wait_until_complete() {
  while true; do
    STATUS=$(aws cloudformation describe-stacks --stack-name $CLOUDFORMATION_STACK_NAME | jq -r .Stacks[0].StackStatus)
    if [[ ${PIPESTATUS[0]} != 0 ]]; then
      echo "failed to describe stacks"
      exit 1
    fi
    if [[ $STATUS = "ROLLBACK_COMPLETE" ]]; then
      echo "Create failed.  Check event logs in web console."
      exit 1
    fi
    if [[ $STATUS = "UPDATE_ROLLBACK_COMPLETE" ]]; then
      echo "Update failed.  Check event logs in web console."
      exit 1
    fi

    if [[ $STATUS = "CREATE_COMPLETE" ]] || [[ $STATUS = "UPDATE_COMPLETE" ]]; then
      echo "$STATUS"
      return
    fi

    echo "Waiting for stack updates to complete.  Current status is: $STATUS"
    sleep 10
  done
}

function idempotent_stack_update() {
  if aws cloudformation update-stack \
     --stack-name "$CLOUDFORMATION_STACK_NAME" \
     --template-body "file:///$THISDIR/cloudformation.json" 2> /tmp/update-result \
     --parameters ParameterKey=HostedZone,ParameterValue="$HOSTED_ZONE_NAME" ParameterKey=SSLCertARN,ParameterValue="$SSL_CERT_ARN" ParameterKey=NatAMI,ParameterValue="$NAT_AMI"
  then
    echo "Stack update started."
    wait_until_complete
    return
  fi

  if grep "No updates are to be performed" /tmp/update-result > /dev/null
  then
    echo "No updates required.  OK"
    return
  else
    echo "Error updating stack"
    cat /tmp/update-result
    exit 1
  fi
}

aws --version
aws configure list

THISDIR=$(cd $(dirname $0) && pwd)

echo -n "stack exists? "
if stack_exists "$CLOUDFORMATION_STACK_NAME"; then
  echo "yes; will update stack"
  idempotent_stack_update
else
  echo "no; will create new stack"
  if ! aws cloudformation create-stack \
     --stack-name "$CLOUDFORMATION_STACK_NAME" \
     --template-body "file:///$THISDIR/cloudformation.json" \
     --parameters ParameterKey=HostedZone,ParameterValue="$HOSTED_ZONE_NAME" ParameterKey=SSLCertARN,ParameterValue="$SSL_CERT_ARN" ParameterKey=NatAMI,ParameterValue="$NAT_AMI"
  then
    echo "failed to create new stack"
    exit 1
  fi

  wait_until_complete
fi
