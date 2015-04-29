# Manual Install for Lattice

## Prerequisites

* Ubuntu-based Linux install
* No previous lattice installations already running

## Installation

1) 
Download the tarball using appropriate link.

Type | Link
------|--------
Release | [https://lattice.s3.amazonaws.com/releases/latest/lattice.tgz](https://lattice.s3.amazonaws.com/releases/latest/lattice.tgz)
Unstable | [https://lattice.s3.amazonaws.com/unstable/latest/lattice.tgz](https://lattice.s3.amazonaws.com/unstable/latest/lattice.tgz)

```bash
$ curl <lattice_download_url> -o ~/lattice.tgz
```

2)
Unpack installer script from `lattice.tgz`

```bash
$ tar xzf ~/lattice.tgz lattice-build/scripts/install-from-tar --strip-components 2 
```

3)
Populate the `lattice-environment` file

```bash
$ mkdir -p /var/lattice/setup
$ sudo tee /var/lattice/setup/lattice-environment >/dev/null <<EOF
CONSUL_SERVER_IP=<system_ip>
SYSTEM_DOMAIN=<system_ip>.xip.io
LATTICE_CELL_ID=cell-01
GARDEN_EXTERNAL_IP=<system_ip>
EOF
```

4) 
Run the installer script

```bash
$ ~/install-from-tar collocated ~/lattice.tgz
```

Note: The lattice cluster can take some time to become available.  Please `ltc target` the new cluster and try `ltc test -v` until the cluster is healthy.