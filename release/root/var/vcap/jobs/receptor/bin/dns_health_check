#!/bin/bash

nc -z 127.0.0.1 8888

# http://www.consul.io/docs/agent/checks.html
if [ $? -ne 0 ]; then
  exit 2 # Exit higher than 1 to ensure service is registered as 'critical'
fi
