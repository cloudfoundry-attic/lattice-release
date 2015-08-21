# Lattice: Run Containerized Workloads

<table>
  <tr>
    <td>
      <a href="http://lattice.cf"><img src="https://raw.githubusercontent.com/cloudfoundry-incubator/lattice/master/docs/logos/lattice.png" align="left" width="200" ></a>
    </td>
    <td>
      Website: <a href="http://lattice.cf">http://lattice.cf</a><br>
      Mailing List: <a href="https://lists.cloudfoundry.org/mailman/listinfo/cf-lattice">Subscribe</a><br>
      Archives: [ <a href="http://cf-lattice.70370.x6.nabble.com/">Nabble</a> | <a href="https://groups.google.com/a/cloudfoundry.org/forum/#!forum/lattice">Google Groups</a> ]
    </td>
  </tr>
</table>

Lattice is an open source project for running containerized workloads on a cluster. Lattice bundles up http load-balancing, a cluster scheduler, log aggregation/streaming and health management into an easy-to-deploy and easy-to-use package.

Lattice is based on a number of open source [Cloud Foundry](http://cloudfoundry.org) components:

- [Diego](https://github.com/cloudfoundry-incubator/diego-design-notes) schedules and monitors containerized workloads
- [Loggregator](https://github.com/cloudfoundry/loggregator) aggregates and streams application logs
- [Gorouter](https://github.com/cloudfoundry/gorouter) provides http load-balancing

## Deploy Lattice

A [local deployment](#local-deployment) of Lattice can be launched with Vagrant.

A scalable [cluster deployment](#clustered-deployment) of Lattice can be launched with Terraform.  We currently support [AWS](terraform/aws/README.md), [DigitalOcean](terraform/digitalocean/README.md), [Google Cloud](terraform/google/README.md) and [Openstack](terraform/openstack/README.md).

## Use Lattice

The [Lattice CLI `ltc`](https://github.com/cloudfoundry-incubator/lattice/tree/master/ltc) provides a command line interface for launching docker-based applications.

More complex workloads can be constructed and submitted directly to Lattice's Receptor API which is fully documented [here](https://github.com/cloudfoundry-incubator/receptor/blob/master/doc/README.md).

# Local Deployment

## Launching with Vagrant

Make sure you have [Vagrant](https://vagrantup.com/) 1.6+ installed, download the Lattice bundle from the [latest release](https://github.com/cloudfoundry-incubator/lattice/releases/latest), and run `vagrant up`:

```bash
unzip lattice-bundle-VERSION-PLATFORM.zip
cd lattice-bundle-VERSION-PLATFORM/vagrant
vagrant up
```

This spins up a virtual environment that is accessible at `192.168.11.11`.

Use the [Lattice CLI](https://github.com/cloudfoundry-incubator/lattice/tree/master/ltc) to target Lattice:

```bash
ltc target 192.168.11.11.xip.io
```

> NOTE: Ubuntu 14.04 LTS does not install a compatible version of vagrant by default. You can upgrade the version that you get out of the box by downloading the `.deb` file from [Vagrant](http://www.vagrantup.com/downloads.html).

## Using Different Providers

You can do this with either VirtualBox or VMware Fusion (version 7 or later):

Virtualbox:

```bash
vagrant up --provider virtualbox
```

VMware Fusion:
```bash
vagrant up --provider vmware_fusion
```

### Networking Conflicts

If you are trying to run both the VirtualBox and VMWare providers on the same machine,
you'll need to run them on different private networks (subnets) that do not conflict.

Set the System IP to an address that does not conflict with the host networking configuration by passing the
`LATTICE_SYSTEM_IP` environment variable to the vagrant up command:

```bash
LATTICE_SYSTEM_IP=192.168.80.100 vagrant up
ltc target 192.168.80.100.xip.io
```

## Miscellaneous

### Nightly builds

The latest unsupported nightly build is available [for Linux](http://lattice.s3.amazonaws.com/nightly/lattice-bundle-latest-linux.zip) and [for OS X](http://lattice.s3.amazonaws.com/nightly/lattice-bundle-latest-osx.zip).

### Building Lattice from source

To build Lattice from source and deploy using Vagrant:

```bash
    $ git clone git@github.com:cloudfoundry-incubator/lattice.git
    $ cd lattice
    $ development/setup && development/build && development/run
    $ source development/env
```

> More information on developing for Lattice can be found on the [development readme](development/README.md).

### Updating

Currently, Lattice does not support updating via provision. To update, you have to destroy the box and bring it back up with a new `Vagrantfile`.
If you have copied a new `Vagrantfile` into an existing directory, make sure to remove any outdated `lattice.tgz` present in that directory.

### Manual install of Lattice

Follow these [instructions](docs/manual-install.md) to install a co-located Lattice cluster to a server that's already deployed. (e.g., vSphere)  

### Proxy configuration

Install the `vagrant-proxyconf` plugin as follows:

```bash
vagrant plugin install vagrant-proxyconf
```

Then proceed with `http_proxy=http://PROXY_IP:PROXY_PORT vagrant up`.

## Troubleshooting

### No such host errors

DNS resolution for `xip.io` addresses can sometimes be flaky, resulting in errors such as the following:

```bash
 ltc target 192.168.11.11.xip.io
 Error verifying target: Get http://receptor.192.168.11.11.xip.io/v1/desired_lrps:
 dial tcp: lookup receptor.192.168.11.11.xip.io: no such host
```

_Resolution Steps_

1. Follow [these instructions](https://support.apple.com/en-us/HT202516) to reset the DNS cache in OS X.  There have been several reported [issues](http://arstechnica.com/apple/2015/01/why-dns-in-os-x-10-10-is-broken-and-what-you-can-do-to-fix-it/) with DNS resolution on OS X, specifically on Yosemite, insofar as the latest beta build of OS X 10.10.4 has [replaced `discoveryd` with `mDNSResponder`](http://arstechnica.com/apple/2015/05/new-os-x-beta-dumps-discoveryd-restores-mdnsresponder-to-fix-dns-bugs/).

1. Check your networking DNS settings. Local "forwarding DNS" servers provided by some home routers can have trouble resolving `xip.io` addresses. Try setting your DNS to point to your real upstream DNS servers, or alternatively try using [Google DNS](https://developers.google.com/speed/public-dns/) by using `8.8.8.8` and/or `8.8.4.4`.

1. If the above steps don't work (or if you must use a DNS server that doesn't work with `xip.io`), our recommended alternative is to follow the [dnsmasq instructions](docs/dnsmasq-readme.md), pass the `LATTICE_SYSTEM_DOMAIN` environment variable to the vagrant up command, and target using `lattice.dev` instead of `192.168.11.11.xip.io` to point to the cluster, as follows:

```
LATTICE_SYSTEM_DOMAIN=lattice.dev vagrant up
ltc target lattice.dev
```

> `dnsmasq` is currently only supported for **vagrant** deployments.

### Miscellaneous

If you have trouble running `vagrant up --provider virtualbox` with the error

```
default: Warning: Remote connection disconnect. Retrying...
default: Warning: Authentication failure. Retrying...
...
```

try upgrading to the latest VirtualBox.

# Clustered Deployment

This repository contains several [Terraform](https://www.terraform.io/) templates to help you deploy on your choice of IaaS.  To deploy Lattice in this way you will need:

* Credentials for your choice of IaaS
* [Terraform](https://www.terraform.io/intro/getting-started/install.html)

  Lattice | Compatible Versions
  --------|-----------------------
  v0.3.3  | Terraform 0.6.2+
  v0.3.0  | Terraform 0.6.1
  v0.2.7  | Terraform 0.6.1
  v0.2.6  | Terraform 0.5.0-0.5.3
  v0.2.5  | Terraform 0.5.0-0.5.3
  v0.2.4  | Terraform 0.4.1-0.4.2
  v0.2.3  | Terraform 0.3.5-0.3.7

## Deploying

Here are some step-by-step instructions for deploying a Lattice cluster via Terraform:

1. Download a lattice bundle from the [latest release](https://github.com/cloudfoundry-incubator/lattice/releases/latest).
2. Unzip the bundle, and switch to the directory for your intended provider:

  ```bash
  unzip lattice-bundle-VERSION-PLATFORM.zip
  cd lattice-bundle-VERSION-PLATFORM/terraform/PROVIDER
  ```

4. Update the `lattice.<provider>.tf` by filling in the values for the variables. Instructions for each supported platform are here:
  - [Amazon Web Services](terraform/aws/README.md#configure)
  - [DigitalOcean](terraform/digitalocean/README.md#configure)
  - [Google Cloud](terraform/google/README.md#configure)
  - [Openstack](terraform/openstack/README.md#configure)
      - Note: This is a community-supplied platform. It is not presently supported by the project maintainers.
5. Run the following commands in the folder containing the `lattice.<platform>.tf` file

  ```bash
  terraform get -update
  terraform apply
  ```

  This will deploy the cluster.

Upon success, terraform will print the Lattice target:

```
Outputs:

  lattice_target = x.x.x.x.xip.io
  lattice_username = xxxxxxxx
  lattice_password = xxxxxxxx
```

which you can use with the Lattice CLI to `ltc target x.x.x.x.xip.io`.

Terraform will generate a `terraform.tfstate` file.  This file describes the cluster that was built - keep it around in order to modify/tear down the cluster.

## Destroying

To destroy the cluster go to the folder containing the `terraform.tfstate` file and run:

```bash
terraform destroy
```

## Updating

To update the cluster, you must destroy your existing cluster and begin the deployment process again using a new lattice bundle.

# Contributing

Everyone is encouraged to help improve this project.

Please submit pull requests against the **develop branch**. Only the continuous integration system commits to master.

Here are some ways *you* can contribute:

* by using alpha, beta, and prerelease versions
* by reporting bugs
* by suggesting new features
* by writing or editing documentation
* by writing specifications
* by writing code (**no patch is too small**: fix typos, add comments, clean up inconsistent whitespace)
* by refactoring code
* by closing [issues](https://github.com/cloudfoundry-incubator/lattice/issues)
* by reviewing patches

Also see the [Development Readme](development/README.md)

## Submitting an Issue
We use the [GitHub issue tracker](https://github.com/cloudfoundry-incubator/lattice/issues) to track bugs and features.
Before submitting a bug report or feature request, check to make sure it hasn't already been submitted.
You can indicate support for an existing issue by voting it up.
When submitting a bug report, please include a [Gist](http://gist.github.com/) that includes a stack trace and any
details that may be necessary to reproduce the bug including the Lattice version.

## Submitting a Pull Request

1. Propose a change by opening an issue.
2. Fork the project.
3. Create a topic branch.
4. Implement your feature or bug fix.
5. Commit and push your changes.
6. Submit a pull request.

## Choosing Stories From our [Tracker](https://www.pivotaltracker.com/n/projects/1183596)

[Pivotal Tracker](http://www.pivotaltracker.com/) is the way that Cloud Foundry projects organize and prioritize work and releases. With Tracker, work is organized into stories that are actionable and have been prioritized.  The team typically works on stories found in the Current and Backlog columns.

Stories not (yet) prioritized are kept in the Icebox. The Icebox is a grab-bag of feature requests, GitHub Issues, or partially-developed ideas. Stories in the Icebox may never be prioritized, and may not always be maintained in the same disciplined manner as the Backlog.

Periodically, the Lattice team goes through through both the Backlog and the Icebox, and tags stories using the 'community' label.   These are stories that are particularly well-suited to community contribution, and make excellent candidates for people to work on.

# Copyright

See [LICENSE](docs/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
