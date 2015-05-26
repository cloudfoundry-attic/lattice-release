# Manual Install for Lattice

Lattice supports two different deployment methods: Vagrant, for the local host, and Terraform which automates the process of deploying a Lattice cluster to public clouds. Automated deployment to private clouds is not supported. However, deployment can be done manually. Follow these steps to deploy Lattice to a single pre-provisioned VM or host.

Instructions to deploy a cluster will come in a future release.

## Prerequisites

* Ubuntu-based Linux install
  - At least 12GB of raw disk provisioned for the VM
  - Access to the Internet (to download packages)
* No previous lattice installations already running
* Following packages installed via `apt-get`:
  - `curl`
  - `gcc`
  - `make`
  - `quota`
  - `linux-image-extra-$(uname -r)`

## Installation

1) 
Download the tarball using appropriate link.

- Release - [https://lattice.s3.amazonaws.com/releases/latest/lattice.tgz](https://lattice.s3.amazonaws.com/releases/latest/lattice.tgz)
- Unstable - [https://lattice.s3.amazonaws.com/unstable/latest/lattice.tgz](https://lattice.s3.amazonaws.com/unstable/latest/lattice.tgz)


```bash
$ curl <lattice_download_url> -o lattice.tgz
```

2)
Unpack installer script from `lattice.tgz`

```bash
$ tar xzf lattice.tgz --strip-components 2 lattice-build/scripts/install-from-tar
```

3)
Populate the `lattice-environment` file

```bash
$ sudo mkdir -p /var/lattice/setup
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
$ sudo ./install-from-tar collocated lattice.tgz
```

Note: The lattice cluster can take some time to become available.  Please `ltc target` the new cluster and try `ltc test -v` until the cluster is healthy.
