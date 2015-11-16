#!/bin/bash

set -e

if [[ ! -f $1 ]]; then
  echo "Usage:"
  echo -e "\t$0 /path/to/lattice.tgz"
  exit 1
fi

export $(cat /var/lattice/setup)

/etc/init.d/lighttpd stop >/dev/null
tar xzf "$1" --strip-components 1 --keep-directory-symlink --no-overwrite-dir --no-same-owner -C / -- brain

# Configure proxy support
mkdir -p /var/lattice/lighttpd
if [[ -f /var/lattice/proxy ]]; then
  export $(cat /var/lattice/proxy)
  cat > /var/lattice/lighttpd/proxyconf.json <<JSON
  {
      "http_proxy": $(echo $http_proxy | jq -R .),
      "https_proxy": $(echo $https_proxy | jq -R .),
      "no_proxy": $(echo $no_proxy | jq -R .)
  }
JSON
fi

# Configure and secure lighttpd
mkdir -p /var/lattice/lighttpd/blobs
chown -R www-data:www-data /var/lattice/lighttpd
lighttpd_password=$(openssl passwd -crypt "$PASSWORD")
echo $USERNAME:$lighttpd_password > /var/lattice/lighttpd.user

# Configure receptor
new_ip=$(ip route get 1 | awk '{print $NF;exit}')
receptor_ctl=/var/vcap/jobs/receptor/bin/receptor_ctl
sed -i "s/replace-receptor-ip/$new_ip/g" "$receptor_ctl"
sed -i "s/replace-receptor-domain/$DOMAIN/g" "$receptor_ctl"
sed -i "s/replace-receptor-username/$USERNAME/g" "$receptor_ctl"
sed -i "s/replace-receptor-password/$PASSWORD/g" "$receptor_ctl"
ln -sf /var/vcap/jobs/receptor/monit /var/vcap/monit/job/0019_receptor.monitrc

# Configure loggregator
sed -i 's%/var/vcap/packages/loggregator_trafficcontroller/trafficcontroller%& --disableAccessControl%' \
  /var/vcap/jobs/loggregator_trafficcontroller/bin/loggregator_trafficcontroller_ctl

/etc/init.d/lighttpd start >/dev/null
