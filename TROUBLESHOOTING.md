# Troubleshooting

## No such host errors

DNS resolution for `xip.io` addresses can sometimes be flaky, resulting in errors such as the following:

```bash
 ltc target 192.168.11.11.xip.io
 Error verifying target: Get http://receptor.192.168.11.11.xip.io/v1/desired_lrps:
 dial tcp: lookup receptor.192.168.11.11.xip.io: no such host
```

_Resolution Steps_

1. Follow [these instructions](https://support.apple.com/en-us/HT202516) to reset the DNS cache in OS X.  There have been several reported [issues](http://arstechnica.com/apple/2015/01/why-dns-in-os-x-10-10-is-broken-and-what-you-can-do-to-fix-it/) with DNS resolution on OS X, specifically on Yosemite, insofar as the latest beta build of OS X 10.10.4 has [replaced `discoveryd` with `mDNSResponder`](http://arstechnica.com/apple/2015/05/new-os-x-beta-dumps-discoveryd-restores-mdnsresponder-to-fix-dns-bugs/).

1. Check your networking DNS settings. Local "forwarding DNS" servers provided by some home routers can have trouble resolving `xip.io` addresses. Try setting your DNS to point to your real upstream DNS servers, or alternatively try using [Google DNS](https://developers.google.com/speed/public-dns/) by using `8.8.8.8` and/or `8.8.4.4`.

1. If the above steps don't work (or if you must use a DNS server that doesn't work with `xip.io`), our recommended alternative is to follow the [dnsmasq instructions](http://lattice.cf/docs/dnsmasq-readme), pass the `LATTICE_SYSTEM_DOMAIN` environment variable to the vagrant up command, and target using `lattice.dev` instead of `192.168.11.11.xip.io` to point to the cluster, as follows:

```
LATTICE_SYSTEM_DOMAIN=lattice.dev vagrant up
ltc target lattice.dev
```

> `dnsmasq` is currently only supported for **vagrant** deployments.

## Vagrant IP conflict errors

The below errors can come from having multiple vagrant instances using the same IP address (e.g., 192.168.11.11).  

```bash
$ ltc target 192.168.11.11.xip.io
Error connecting to the receptor. Make sure your lattice target is set, and that lattice is up and running.
  Underlying error: Get http://receptor.192.168.11.11.xip.io/v1/desired_lrps: read tcp 192.168.11.11:80: connection reset by peer

$ ltc target 192.168.11.11.xip.io
Error connecting to the receptor. Make sure your lattice target is set, and that lattice is up and running.
  Underlying error: Get http://receptor.192.168.11.11.xip.io/v1/desired_lrps: use of closed network connection  

$ ltc target 192.168.11.11.xip.io
Error verifying target: Get http://receptor.192.168.11.11.xip.io/v1/desired_lrps: net/http: transport closed before response was received
``` 

To check whether multiple VMs might have an IP conflict, run the following:

```bash
$ vagrant global-status
id       name    provider   state   directory
----------------------------------------------------------------------------------------------------------------
fb69d90  default virtualbox running /Users/user/workspace/lattice
4debe83  default virtualbox running /Users/user/workspace/lattice-bundle-v0.4.0-osx/vagrant
```

You can then destroy the appropriate instance with:

```bash
$ cd </path/to/vagrant-directory>
$ vagrant destroy
```

## Miscellaneous

If you have trouble running `vagrant up --provider virtualbox` with the error

```
default: Warning: Remote connection disconnect. Retrying...
default: Warning: Authentication failure. Retrying...
...
```

try upgrading to the latest VirtualBox.

