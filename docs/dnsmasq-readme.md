#DNSMasq Configuration Readme

## Installation 

* [Source link](http://passingcuriosity.com/2013/dnsmasq-dev-osx/)

1. Use Homebrew to install `dnsmasq`
```bash
# Update your homebrew installation
brew up
# Install dnsmasq
brew install dnsmasq
```

2. Use the commands provided by `brew info dnsmasq` to configure and start the service:  
* As of dnsmasq v2.72, `brew info dnsmasq` returned:
```
$ brew info dnsmasq
dnsmasq: stable 2.72 (bottled)
http://www.thekelleys.org.uk/dnsmasq/doc.html
/usr/local/Cellar/dnsmasq/2.72 (7 files, 496K) *
  Built from source
From: https://github.com/Homebrew/homebrew/blob/master/Library/Formula/dnsmasq.rb
...
==> Caveats
To configure dnsmasq, copy the example configuration to /usr/local/etc/dnsmasq.conf
and edit to taste.

  cp /usr/local/opt/dnsmasq/dnsmasq.conf.example /usr/local/etc/dnsmasq.conf

To have launchd start dnsmasq at startup:
    sudo cp -fv /usr/local/opt/dnsmasq/*.plist /Library/LaunchDaemons
    sudo chown root /Library/LaunchDaemons/homebrew.mxcl.dnsmasq.plist
Then to load dnsmasq now:
    sudo launchctl load /Library/LaunchDaemons/homebrew.mxcl.dnsmasq.plist
```

## Configuration

### Set up dnsmasq to resolve `lattice.dev`

1. Append line to `/usr/local/etc/dnsmasq.conf` 
```
address=/lattice.dev/<LATTICE_SYSTEM_IP> # i.e., 192.168.11.11
```

2. Restart dnsmasq service
```bash
sudo launchctl stop homebrew.mxcl.dnsmasq
sudo launchctl start homebrew.mxcl.dnsmasq
```

### Configure workstation to use dnsmasq resolver for `lattice.dev`

1. Create `/etc/resolver` folder
```bash
sudo mkdir /etc/resolver
```
2. Define resolver for `lattice.dev`
```bash
sudo tee /etc/resolver/lattice.dev >/dev/null <<EOF
nameserver 127.0.0.1
EOF
```

## Starting Lattice cluster with alternate name

1. Set `LATTICE_SYSTEM_DOMAIN` environment variable during `vagrant up`:
```
LATTICE_SYSTEM_DOMAIN=lattice.dev vagrant up --provider=<PROVIDER> 
```