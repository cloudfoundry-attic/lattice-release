#!/bin/bash

set -e

export $(cat /var/lattice/setup)

# Provide Consul with the Lattice brain IP

agent_ctl=/var/vcap/jobs/consul_agent/bin/agent_ctl
sed -i 's/expected=0/expected=1/' "$agent_ctl"
sed -i 's/consul_join=""/consul_join="'$BRAIN_IP'"/' "$agent_ctl"
sed -i 's/consul_server_ips=""/consul_server_ips="'$BRAIN_IP'"/' "$agent_ctl"

consul_config=/var/vcap/jobs/consul_agent/config/config.json
consul_config_json=$(cat "$consul_config")
echo $consul_config_json | jq '. + {"retry_join": ["'$BRAIN_IP'"]}' > "$consul_config"

# Replace placeholder IP with Lattice brain IP

job_files=$(find /var/vcap/jobs/*/ -type f)
old_brain_ip=$(cat /var/lattice/brain_ip)
perl -p -i -e "s/\\Q$old_brain_ip\\E/$BRAIN_IP/g" $job_files
