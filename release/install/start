#!/bin/bash

set -e

rm -f /etc/init/runsvdir.override
start runsvdir >/dev/null

echo "Waiting for services to start..."
monit_summary="/var/vcap/bosh/bin/monit summary"
started_services() { $monit_summary 2>/dev/null | grep -E '(running|accessible)' | wc -l; }
total_services() { $monit_summary 2>/dev/null | grep -E '^(Process|File|System)' | wc -l; }
while [[ $(total_services) = 0 ]] || [[ $(started_services) -lt $(total_services) ]]; do
  sleep 1
done
