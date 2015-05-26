# DNSMasq Configuration on OSX

Many Lattice services run over HTTP. Via the [Gorouter](https://github.com/cloudfoundry/gorouter), they share the same IP address. They are distinguished based on which hostname they've been accessed by. That's why many Lattice examples require the use of `.xip.io` instead of the raw IP address. That way, the client correctly communicates the domain name to the service regardless of how many services share that IP address.

We've taken to using [dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) to avoid the use of `.xip.io` above. Dnsmaq is a useful tool to provide a private, local DNS server that can be configured to return the IP address of your Lattice installment. In this document, the examples use the [dnsmasq package for OSX](http://passingcuriosity.com/2013/dnsmasq-dev-osx/). If you are not on OSX, you can use the [source distribution](http://www.thekelleys.org.uk/dnsmasq/).

After following these instructions, wherever you see `servicename.IP-ADDRESS.xip.io` you can simply use `servicename.lattice.dev`. So, `ltc target 192.168.11.11.xip.io` becomes `ltc target lattice.dev`.

## Installation 

- Use Homebrew to install `dnsmasq`:

```bash
# Update your homebrew installation
$ brew up
# Install dnsmasq
$ brew install dnsmasq
```

- Use the commands provided by `brew info dnsmasq` to configure and start the service.<br>
  As of dnsmasq v2.72, `brew info dnsmasq` returned:

```bash
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

### Set up dnsmasq to resolve lattice.dev
- Using your favorite text editor, append the following line to `/usr/local/etc/dnsmasq.conf`. Make sure to replace `<LATTICE_SYSTEM_IP>` with your Lattice target IP address.

```bash
address=/lattice.dev/<LATTICE_SYSTEM_IP> # i.e., 192.168.11.11
```

- Restart the dnsmasq service

```bash
$ sudo launchctl stop homebrew.mxcl.dnsmasq
$ sudo launchctl start homebrew.mxcl.dnsmasq
```

### Configure your workstation to use the dnsmasq resolver for lattice.dev
- Create `/etc/resolver` folder

```bash
$ sudo mkdir /etc/resolver
```

- Create a file that defines the resolver for `lattice.dev`

```bash
$ sudo tee /etc/resolver/lattice.dev >/dev/null <<EOF
nameserver 127.0.0.1
EOF
```

## Starting Lattice cluster with alternate name
- Set `LATTICE_SYSTEM_DOMAIN` environment variable during `vagrant up`:

```bash
LATTICE_SYSTEM_DOMAIN=lattice.dev vagrant up --provider=<PROVIDER> 
```

### Validating your dnsmasq setup
Here's how you can prove that you're set up to redirect requests for the lattice.dev domain, as well as make sure that regular DNS resolution has not been affected.

```bash
$ host www.lattice.dev 127.0.0.1
Using domain server:
Name: 127.0.0.1
Address: 127.0.0.1#53
Aliases: 

www.lattice.dev has address 192.168.11.11
```

```bash
$ host www.yahoo.com 127.0.0.1
Using domain server:
Name: 127.0.0.1
Address: 127.0.0.1#53
Aliases: 

www.yahoo.com is an alias for fd-fp3.wg1.b.yahoo.com.
fd-fp3.wg1.b.yahoo.com has address 206.190.36.45
fd-fp3.wg1.b.yahoo.com has address 206.190.36.105
fd-fp3.wg1.b.yahoo.com has IPv6 address 2001:4998:c:a06::2:4008
```
