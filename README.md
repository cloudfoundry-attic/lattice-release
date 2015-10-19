# Cloud Foundry Lattice

<table width="100%" border="0">
  <tr>
    <td>
      <a href="http://lattice.cf"><img src="https://raw.githubusercontent.com/cloudfoundry-incubator/lattice/master/lattice.png" align="left" width="300px" ></a>
    </td>
    <td>Lattice is an open source project for running containerized workloads on a cluster. Lattice bundles up http load-balancing, a cluster scheduler, log aggregation/streaming and health management into an easy-to-deploy and easy-to-use package.
    </td>
  </tr>
</table>

[ [Website](http://lattice.cf) | [Latest Release](https://github.com/cloudfoundry-incubator/lattice/releases/latest) | [Nightly Builds](https://lattice.s3.amazonaws.com/nightly/index.html) ]

## Deploy Lattice with Vagrant

A collocated deployment of Lattice can be launched locally with [Vagrant](https://vagrantup.com/). You will need:

* A Lattice bundle from the [latest release](https://github.com/cloudfoundry-incubator/lattice/releases/latest) or [nightly builds](https://lattice.s3.amazonaws.com/nightly/index.html)
* [Vagrant](https://vagrantup.com/) 1.6+ installed 

> NOTE: Ubuntu 14.04 LTS does not install a compatible version of Vagrant by default. You can upgrade the version that you get out of the box by downloading the `.deb` file from [Vagrant](http://www.vagrantup.com/downloads.html).

##### Spin up a virtual environment

Unzip the Lattice bundle, and switch to the vagrant directory

```bash
unzip lattice-bundle-VERSION-PLATFORM.zip
cd lattice-bundle-VERSION-PLATFORM/vagrant
vagrant up --provider virtualbox
```

This spins up a virtual environment that is accessible at `192.168.11.11`

##### Install the Lattice CLI

If you're running Linux: `curl -O http://receptor.192.168.11.11.xip.io/v1/sync/linux/ltc`

If you're running OS X: `curl -O http://receptor.192.168.11.11.xip.io/v1/sync/osx/ltc`

```bash
chmod a+x ltc
./ltc target 192.168.11.11.xip.io
./ltc -v
```

For more information visit [Lattice CLI](https://github.com/cloudfoundry-incubator/lattice/blob/master/ltc/README.md)

##### Use the Lattice CLI to target Lattice

```bash
cd lattice-bundle-VERSION-PLATFORM
./ltc target 192.168.11.11.xip.io
```

## Deploy Lattice with Terraform

A scalable cluster deployment of Lattice can be launched on [AWS](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/aws/README.md) with [Terraform](https://www.terraform.io). You will need:

* An [Amazon Web Services account](http://aws.amazon.com/)
* [AWS Access and Secret Access Keys](http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSGettingStartedGuide/AWSCredentials.html)
* [AWS EC2 Key Pairs](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)
* [Terraform 0.6.2+](https://www.terraform.io/intro/getting-started/install.html)
* A Lattice bundle from the [latest release](https://github.com/cloudfoundry-incubator/lattice/releases/latest) or the [nightly builds](https://lattice.s3.amazonaws.com/nightly/index.html) page

##### Configure your virtual environment

Unzip the Lattice bundle, and switch to the terraform/aws directory

```bash
unzip lattice-bundle-VERSION-PLATFORM.zip
cd lattice-bundle-VERSION-PLATFORM/terraform/aws
```

Update the `lattice.aws.tf` by [filling in the values for the variables](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/aws/README.md#configure).

##### Deploy the cluster to AWS

From the folder containing the `lattice.aws.tf`, run these commands to deploy your cluster 

```bash
terraform get -update
terraform apply
```

Terraform will generate a `terraform.tfstate` file.  This file describes the cluster that was built - keep it around in order to modify/tear down the cluster.

##### Install `ltc` (the Lattice CLI)

After a successful deployment Terraform will print the Lattice target and Lattice user information. Refer to the `lattice_target = x.x.x.x.xip.io` output line to find the address of your cluster.

If you're running Linux: `curl -O http://receptor.x.x.x.x.xip.io/v1/sync/linux/ltc`

If you're running OS X: `curl -O http://receptor.x.x.x.x.xip.io/v1/sync/osx/ltc`

```bash
chmod a+x ltc
./ltc target x.x.x.x.xip.io
./ltc -v
```

## Development

> NOTE: These instructions are for people contributing code to Lattice. If you want to install a Lattice release, see above.

To develop lattice release you will need to have the following tools installed:

- vagrant
- packer
- virtualbox
- direnv _(optional)_

### Clone the Lattice source

```bash
git clone --recursive https://github.com/cloudfoundry-incubator/lattice-release.git
cd lattice-release
```

### Build Lattice

Setup your shell for building Lattice:

```bash
# in lattice-release
direnv allow
# or
source .envrc
```

If you want to build and deploy Lattice on AWS using Vagrant, you'll need to add your AWS credentials to the environment:

```bash
export AWS_ACCESS_KEY_ID=<...>
export AWS_SECRET_ACCESS_KEY=<...>
export AWS_SSH_PRIVATE_KEY_NAME=<...> # name of the remote SSH key in AWS
export AWS_SSH_PRIVATE_KEY_PATH=<...> # path to the local SSH key
export AWS_INSTANCE_NAME=<...> # optional
```

The first time you build Lattice, you'll need to create a local Vagrant box containing Diego. You can skip this step for subsequent Lattice builds unless the Diego release has changed.

```bash
# in lattice-release
bundle
cd vagrant

./build -only=virtualbox-iso
vagrant box add --force lattice-virtualbox-v0.box --name lattice/collocated
# or 
./build -only=vmware-iso
vagrant box add --force lattice-vmware-v0.box --name lattice/collocated
# or 
./build -only=amazon-ebs # NOTE: Requires AWS credentials in the environment
vagrant box add --force lattice-aws-v0.box --name lattice/collocated

cd ..
```

Finally, build the Lattice tarball:

```bash
# in lattice-release
./release/build vagrant/lattice.tgz
```

### Deploy Lattice

Once you have a Lattice tarball, use `vagrant` to deploy Lattice:

```bash
# in lattice-release
cd vagrant
vagrant up --provider=virtualbox
# or
vagrant up --provider=vmware_fusion
# or
vagrant up --provider=aws # NOTE: Requires AWS credentials in the environment
```

### Install ltc and test deployed Lattice cluster

Compiling ltc is as simple as using `go install`:

```bash
# in lattice-release
go install github.com/cloudfoundry-incubator/lattice/ltc
```

With a running Lattice cluster, target and run the cluster tests:

```bash
# in lattice-release
ltc target local.lattice.cf
ltc test -v
```

For more information, visit [Lattice CLI](https://github.com/cloudfoundry-incubator/lattice/blob/master/ltc/README.md).
