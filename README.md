# Cloud Foundry Lattice

<table width="100%" border="0">
  <tr>
    <td>
      <a href="http://lattice.cf"><img src="https://raw.githubusercontent.com/cloudfoundry-incubator/lattice-release/master/lattice.png" align="left" width="300px" ></a>
    </td>
    <td>Lattice is an open source project for running containerized workloads on a cluster. Lattice bundles up http load-balancing, a cluster scheduler, log aggregation/streaming and health management into an easy-to-deploy and easy-to-use package.
    </td>
  </tr>
</table>

[ [Website](http://lattice.cf) | [Latest Release](https://github.com/cloudfoundry-incubator/lattice-release/releases/latest) | [Nightly Builds](https://lattice.s3.amazonaws.com/nightly/index.html) ]

## Deploy Lattice with Vagrant

A colocated deployment of Lattice can be launched locally with [Vagrant](https://vagrantup.com/). You will need:

* A Lattice bundle from the [latest release](https://github.com/cloudfoundry-incubator/lattice-release/releases/latest) or [nightly builds](https://lattice.s3.amazonaws.com/nightly/index.html)
* [Vagrant](https://vagrantup.com/) 1.6+ installed

> NOTE: Ubuntu 14.04 LTS does not install a compatible version of Vagrant by default. You can upgrade the version that you get out of the box by downloading the `.deb` file from [Vagrant](http://www.vagrantup.com/downloads.html).

##### Spin up a virtual environment

Unzip the Lattice bundle, and switch to the vagrant directory

```bash
unzip lattice-bundle-VERSION.zip
cd lattice-bundle-VERSION/vagrant
vagrant up --provider virtualbox
```

This spins up a virtual environment that is accessible at `local.lattice.cf`

##### Install `ltc` (the Lattice CLI)

If you're running Linux: `curl -O http://receptor.local.lattice.cf/v1/sync/linux/ltc`

If you're running OS X: `curl -O http://receptor.local.lattice.cf/v1/sync/osx/ltc`

Finally: `chmod +x ltc`

##### Use the Lattice CLI to target Lattice

```bash
./ltc target local.lattice.cf
```

## Deploy Lattice with Terraform

A scalable cluster deployment of Lattice can be launched on Amazon Web Services with [Terraform](https://www.terraform.io). You will need:

* An [Amazon Web Services account](http://aws.amazon.com/)
* [AWS Access and Secret Access Keys](http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSGettingStartedGuide/AWSCredentials.html)
* [AWS EC2 Key Pairs](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)
* [Terraform 0.6.2+](https://www.terraform.io/intro/getting-started/install.html)
* A Lattice bundle from the [latest release](https://github.com/cloudfoundry-incubator/lattice-release/releases/latest) or the [nightly builds](https://lattice.s3.amazonaws.com/nightly/index.html) page

##### Configure your virtual environment

Unzip the Lattice bundle, and switch to the terraform/aws directory

```bash
unzip lattice-bundle-VERSION.zip
cd lattice-bundle-VERSION/terraform/aws
```

Update the `terraform.tfvars` file with your AWS credentials and desired cluster configuration.

##### Deploy the cluster to AWS

```bash
terraform apply
```

Terraform will generate a `terraform.tfstate` file.  This file describes the cluster that was built - keep it around in order to modify/tear down the cluster.

##### Install `ltc` (the Lattice CLI)

After a successful deployment Terraform will print the Lattice target and Lattice user information. Refer to the `target = <lattice target>` output line to find the address of your cluster.

If you're running Linux: `curl -O http://receptor.<lattice target>/v1/sync/linux/ltc`

If you're running OS X: `curl -O http://receptor.<lattice target>/v1/sync/osx/ltc`

Finally: `chmod +x ltc`

##### Use the Lattice CLI to target Lattice

```bash
./ltc target <lattice target>
```

## Development

> NOTE: These instructions are for people contributing code to Lattice. If you only want to deploy Lattice, see above.
> These instructions cover Vagrant/Virtualbox development.
> A similar process can be followed for Vagrant/VMWare, Vagrant/AWS, and Terraform/AWS development. More documentation is forthcoming.

To develop Lattice you will need to have the following tools installed:

- Packer
- Vagrant
- Virtualbox
- Direnv _(optional)_

### Clone the Lattice source

```bash
git clone --recursive https://github.com/cloudfoundry-incubator/lattice-release.git
```

### Build Lattice

Setup your shell for building Lattice:

```bash
cd lattice-release
direnv allow # or: source .envrc
```

#### Building a Lattice Box

If you change any Diego components, you'll need to build a local Vagrant box with your changes.
If you don't plan to change any Diego components, you can update the `box_version` property
in `vagrant/Vagrantfile` to point to a [pre-built Lattice box on Atlas](https://atlas.hashicorp.com/lattice/boxes/colocated)
and skip this step.

```bash
bundle
cd vagrant
./build -only=virtualbox-iso
vagrant box add --force lattice-virtualbox-v0.box --name lattice/colocated
```

#### Building a release of Lattice

```bash
# in lattice-release/vagrant
../release/build lattice.tgz
```

### Deploy Lattice

Once you have a Lattice tarball, use `vagrant` to deploy Lattice:

```bash
# in lattice-release/vagrant
vagrant up --provider=virtualbox
```

### Install ltc

Compiling ltc is as simple as using `go install`:

```bash
go install github.com/cloudfoundry-incubator/ltc
```

### Test the running Lattice Cluster

```bash
ltc target local.lattice.cf
ltc test -v
```

## Contributing

If you are interested in contributing to Lattice, please refer to [CONTRIBUTING](CONTRIBUTING.md).

# Copyright

See [LICENSE](LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
