# Lattice: run containerized workloads on a cluster with ease

Lattice is an open source project for running containerized workloads on a cluster. Lattice bundles up http load-balancing, a cluster scheduler, log aggregation/streaming and health management into an easy-to-deploy and easy-to-use package.

Lattice is based on a number of open source [Cloud Foundry](http://cloudfoundry.org) components:

- [Diego](https://github.com/cloudfoundry-incubator/diego-design-notes) schedules and monitors containerized workloads
- [Doppler](https://github.com/cloudfoundry/loggregator) aggregates and streams application logs
- [Gorouter](https://github.com/cloudfoundry/gorouter) provides http load-balancing

## Deploy Lattice

A [local deployment](#local-deployment) of Lattice can be launched with Vagrant.

A scalable [cluster deployment](#clustered-deployment) of Lattice can be launched with Terraform.  We currently support [AWS](#amazon-web-services), [DigitalOcean](#digitalocean), and [Google Cloud](#google-cloud)

## Use Lattice

The [Lattice CLI `ltc`](https://github.com/cloudfoundry-incubator/lattice/tree/master/ltc) provides a command line interface for launching docker-based applications.

More complex workloads can be constructed and submitted directly to Lattice's Receptor API which is fully documented [here](https://github.com/cloudfoundry-incubator/receptor/blob/master/doc/README.md).

# Local Deployment

## Launching with Vagrant

Make sure you have [Vagrant](https://vagrantup.com/) installed, then:

    $ git clone git@github.com:cloudfoundry-incubator/lattice.git
    $ cd lattice
    $ vagrant up

This spins up a virtual environment that is accessible at `192.168.11.11`.

Use the [Lattice Cli](https://github.com/cloudfoundry-incubator/lattice/tree/master/ltc) to target Lattice:

```
ltc target 192.168.11.11.xip.io
```

## Using Different Providers

You can do this with either VMware Fusion or VirtualBox:

Virtualbox:

     $ vagrant up --provider virtualbox

VMware Fusion:

     $ vagrant up --provider vmware_fusion

### Networking Conflicts

If you are trying to run both the VirtualBox and VMWare providers on the same machine, 
you'll need to run them on different private networks (subnets) that do not conflict.

Set the System IP to an address that does not conflict with the host networking configuration by passing the
LATTICE_SYSTEM_IP environment variable to the vagrant up command:

```
LATTICE_SYSTEM_IP=192.168.80.100 vagrant up
ltc target 192.168.80.100.xip.io
```

## Updating

Currently, Lattice does not support updating via provision. So to update, you have to destroy the box and bring it back up:

     vagrant destroy --force
     git pull
     vagrant up
  
## Troubleshooting

-  xip.io is sometimes flaky, resulting in no such host errors.
-  The alternative that we have found is to use dnsmasq configured to resolve all xip.io addresses to 192.168.11.11.
-  This also requires creating a /etc/resolvers/io file that points to 127.0.0.1. See further instructions [here] (http://passingcuriosity.com/2013/dnsmasq-dev-osx/). 

## Running Vagrant with a custom Lattice tar

By default, `vagrant up` will fetch the latest Lattice binary tarball.  To use a particular tarball:

    VAGRANT_LATTICE_TAR_PATH=/path/to/lattice.tgz vagrant up

# Clustered Deployment

This repository contains several [Terraform](https://www.terraform.io/) templates to help you deploy on your choice of IaaS.  To deploy Lattice in this way you will need:

* [Terraform](https://www.terraform.io/intro/getting-started/install.html) >= 0.3.6 installed on your machine
* Credentials for your choice of IaaS

## Bootstrapping a Clustered Deployment

### [Amazon Web Services](http://aws.amazon.com/):

Create a `lattice.tf` file by downloading the [AWS example file](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/aws/lattice.tf.example):

``` bash
wget --quiet https://raw.githubusercontent.com/cloudfoundry-incubator/lattice/master/terraform/aws/lattice.tf.example -O lattice.tf
```

Update the downloaded file filling the variables according to the [AWS README](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/aws/README.md) file.

### [DigitalOcean](https://www.digitalocean.com):

Create a `lattice.tf` file by downloading the [DigitalOcean example file](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/digitalocean/lattice.tf.example):

``` bash
wget --quiet https://raw.githubusercontent.com/cloudfoundry-incubator/lattice/master/terraform/digitalocean/lattice.tf.example -O lattice.tf
```

Update the downloaded file filling the variables according to the [DigitalOcean README](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/digitalocean/README.md) file.

### [Google Cloud](https://cloud.google.com/):

Create a `lattice.tf` file downloading the [Google Cloud example file](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/google/lattice.tf.example):

``` bash
wget --quiet https://raw.githubusercontent.com/cloudfoundry-incubator/lattice/lattice-terraform/master/google/lattice.tf.example -O lattice.tf
```
Update the downloaded file filling the variables according to the [Google Cloud README](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/google/README.md) file.

## Deploying

Once your `lattice.tf` file is configured:

```
terraform get -update
terraform apply
```

will deploy the cluster.

Upon success, terraform will print the Lattice target:

```
Outputs:

  lattice_target = x.x.x.x.xip.io
  lattice_username = xxxxxxxx
  lattice_password = xxxxxxxx
```

which you can use with the Lattice CLI to `ltc target x.x.x.x.xip.io`.

Terraform will generate a `lattice.tfstate` file.  This file describes the cluster that was built - keep it around in order to modify/tear down the cluster.

## Destroying

To destroy the cluster:

```
terraform destroy
```

# Contributing

In the spirit of [free software](http://www.fsf.org/licensing/essays/free-sw.html), **everyone** is encouraged to help improve this project.

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

Also see the [Development Readme](https://github.com/cloudfoundry-incubator/lattice/tree/master/development-readme.md)

##Development Workflow

Development work should be done on the develop branch.
As a general rule, only CI should commit to master.

## Submitting an Issue
We use the [GitHub issue tracker](https://github.com/cloudfoundry-incubator/lattice/issues) to track bugs and features.
Before submitting a bug report or feature request, check to make sure it hasn't already been submitted.
You can indicate support for an existing issue by voting it up.
When submitting a bug report, please include a [Gist](http://gist.github.com/) that includes a stack trace and any
details that may be necessary to reproduce the bug, including your gem version, Ruby version, and operating system.
Ideally, a bug report should include a pull request with failing specs.

## Submitting a Pull Request

1. Fork the project.
2. Create a topic branch.
3. Implement your feature or bug fix.
4. Commit and push your changes.
5. Submit a pull request.

# Copyright

See [LICENSE](https://github.com/cloudfoundry-incubator/lattice/blob/master/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
