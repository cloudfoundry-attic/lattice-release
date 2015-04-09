# Lattice: Run Containerized Workloads

<table>
  <tr>
    <td>
      <a href="http://lattice.cf"><img src="https://github.com/cloudfoundry-incubator/lattice/raw/develop/logos/lattice.png" align="left" width="200" ></a>
    </td>
    <td>
      Website: <a href="http://lattice.cf">http://lattice.cf</a><br>
      Mailing List: <a href="https://groups.google.com/a/cloudfoundry.org/forum/#!forum/lattice">Google Groups</a>
    </td>
  </tr>
</table>

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

```bash
git clone git@github.com:cloudfoundry-incubator/lattice.git
cd lattice
git checkout <VERSION>
vagrant up
```

This spins up a virtual environment that is accessible at `192.168.11.11`.  Here, `VERSION` refers to the tagged version you wish to deploy.  These tagged versions are known to be stable.

Use the [Lattice CLI](https://github.com/cloudfoundry-incubator/lattice/tree/master/ltc) to target Lattice:

```bash
ltc target 192.168.11.11.xip.io
```

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
LATTICE_SYSTEM_IP environment variable to the vagrant up command:

```bash
LATTICE_SYSTEM_IP=192.168.80.100 vagrant up
ltc target 192.168.80.100.xip.io
```

## Updating

Currently, Lattice does not support updating via provision. So to update, you have to destroy the box and bring it back up:

```bash
 vagrant destroy --force
 git pull
 vagrant up
```

## Troubleshooting

-  xip.io is sometimes flaky, resulting in no such host errors.
-  The alternative that we have found is to use dnsmasq configured to resolve all xip.io addresses to 192.168.11.11.
-  This also requires creating a /etc/resolvers/io file that points to 127.0.0.1. See further instructions [here] (http://passingcuriosity.com/2013/dnsmasq-dev-osx/). 

## Running Vagrant with a custom Lattice tar

By default, `vagrant up` will fetch the latest Lattice binary tarball.  To use a particular tarball:

```bash
VAGRANT_LATTICE_TAR_PATH=/path/to/lattice.tgz vagrant up
```

# Clustered Deployment

This repository contains several [Terraform](https://www.terraform.io/) templates to help you deploy on your choice of IaaS.  To deploy Lattice in this way you will need:

* [Terraform](https://www.terraform.io/intro/getting-started/install.html) == 0.3.7 installed on your machine (Terraform 0.4.0 is currently unsupported, we are looking into fixing this)
* Credentials for your choice of IaaS

## Deploying

Here are some step-by-step instructions for deploying a Lattice cluster via Terraform:

1. Visit the [Lattice GitHub Releases page](https://github.com/cloudfoundry-incubator/lattice/releases#)
2. Select the Lattice version you wish to deploy and download the Terraform example file for your target platform.  The filename will be `lattice.<platform>.tf`
3. Create an empty folder and place the `lattice.<platform>.tf` file in that folder.
4. Update the `lattice.<platform>.tf` by filling in the values for the variables. Instructions for each supported platform are here:
  - [Amazon Web Services](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/aws/README.md#configure)
  - [DigitalOcean](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/digitalocean/README.md#configure)
  - [Google Cloud](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/google/README.md#configure)
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

## Development Workflow

Development work should be done on the develop branch.
As a general rule, only CI should commit to master.

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

# Copyright

See [LICENSE](https://github.com/cloudfoundry-incubator/lattice/blob/master/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
