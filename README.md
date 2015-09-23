# Lattice: Run Containerized Workloads

<table>
  <tr>
    <td>
      <a href="http://lattice.cf"><img src="https://raw.githubusercontent.com/cloudfoundry-incubator/lattice/master/lattice.png" align="left" width="200" ></a>
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

## Get Lattice

Visit [Lattice Releases](https://github.com/cloudfoundry-incubator/lattice/releases) or our [Nightly Bundles](https://lattice.s3.amazonaws.com/nightly/index.html) page to download one of our bundles.  These include both the `ltc` CLI for the appropriate architecture, as well as the Vagrantfile and Terraform examples for a given release or nightly build.

## Deploy Lattice

A [local deployment](#local-deployment) of Lattice can be launched with Vagrant.

A scalable [cluster deployment](#clustered-deployment) of Lattice can be launched with Terraform.  We currently support [AWS](terraform/aws/README.md), [DigitalOcean](terraform/digitalocean/README.md), [Google Cloud](terraform/google/README.md) and [Openstack](terraform/openstack/README.md).

## Use Lattice

The [Lattice CLI `ltc`](https://github.com/cloudfoundry-incubator/lattice/tree/master/ltc) provides a command line interface for launching docker-based applications.

More complex workloads can be constructed and submitted directly to Lattice's Receptor API which is fully documented [here](https://github.com/cloudfoundry-incubator/receptor/blob/master/doc/README.md).

# Local Deployment

## Launching with Vagrant

Make sure you have [Vagrant](https://vagrantup.com/) 1.6+ installed, download the Lattice bundle from the [latest release](https://github.com/cloudfoundry-incubator/lattice/releases/latest) or the [nightly builds](https://lattice.s3.amazonaws.com/nightly/index.html) page, and run `vagrant up`:

```bash
unzip lattice-bundle-VERSION-PLATFORM.zip
cd lattice-bundle-VERSION-PLATFORM/vagrant
vagrant up
```

This spins up a virtual environment that is accessible at `192.168.11.11`.

Use the [Lattice CLI](https://github.com/cloudfoundry-incubator/lattice/tree/master/ltc) to target Lattice:

```bash
cd lattice-bundle-VERSION-PLATFORM
./ltc target 192.168.11.11.xip.io
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

Follow these [instructions](http://lattice.cf/docs/manual-install) to install a co-located Lattice cluster to a server that's already deployed. (e.g., vSphere)

### Proxy configuration

Install the `vagrant-proxyconf` plugin as follows:

```bash
$ vagrant plugin install vagrant-proxyconf
```

Setup your environment in the terminal:

```bash
$ export http_proxy=http://PROXY_IP:PROXY_PORT
$ export https_proxy=http://PROXY_IP:PROXY_PORT
$ export no_proxy=$(ltc target|head -1|awk '{print $2}')
```

Then proceed with `vagrant up`. For `ltc create`, `ltc build-droplet` or `ltc launch-droplet`, you'll need to pass these environment variables into the container. For example:

```bash
$ ltc create -e http_proxy -e https_proxy -e no_proxy lattice-docker-app cloudfoundry/lattice-app
$ ltc build-droplet -e http_proxy -e https_proxy -e no_proxy lattice-droplet go
$ ltc launch-droplet -e http_proxy -e https_proxy -e no_proxy lattice-app lattice-droplet
```


# Clustered Deployment

This repository contains several [Terraform](https://www.terraform.io/) templates to help you deploy on your choice of IaaS.  To deploy Lattice in this way you will need:

* Credentials for your choice of IaaS
* [Terraform](https://www.terraform.io/intro/getting-started/install.html)

  Lattice | Compatible Versions
  --------|--------------------
  v0.3.3+ | Terraform 0.6.2+
  v0.3.0  | Terraform 0.6.1
  v0.2.7  | Terraform 0.6.1

## Deploying

Here are some step-by-step instructions for deploying a Lattice cluster via Terraform:

1. Download a lattice bundle from the [latest release](https://github.com/cloudfoundry-incubator/lattice/releases/latest) or the [nightly builds](https://lattice.s3.amazonaws.com/nightly/index.html) page.
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

Please submit pull requests against the **master branch**. 

Our [Concourse](http://concourse.ci) CI system is available at [ci.lattice.cf](https://ci.lattice.cf).

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

## Troubleshooting

Please read the [troubleshooting guide](https://github.com/cloudfoundry-incubator/lattice/blob/master/TROUBLESHOOTING.md) 

# Copyright

See [LICENSE](LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
