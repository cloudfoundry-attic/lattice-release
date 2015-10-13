# lattice-release

<table width="100%" border="0">
  <tr>
    <td>
      <a href="http://lattice.cf"><img src="https://raw.githubusercontent.com/cloudfoundry-incubator/lattice/master/lattice.png" align="left" width="300px" ></a>
    </td>
    <td>Lattice is an open source project for running containerized workloads on a cluster. Lattice bundles up http load-balancing, a cluster scheduler, log aggregation/streaming and health management into an easy-to-deploy and easy-to-use package.
    </td>
  </tr>
</table>

[Lattice.cf](http://lattice.cf) | [Latest Release](https://github.com/cloudfoundry-incubator/lattice/releases/latest) | [Nightly Builds](https://lattice.s3.amazonaws.com/nightly/index.html)

## Deploy Lattice with Vagrant

A collocated deployment of Lattice can be launched locally with [Vagrant](https://vagrantup.com/). You will need:

* A Lattice bundle from the [latest release](https://github.com/cloudfoundry-incubator/lattice/releases/latest) or [nightly builds](https://lattice.s3.amazonaws.com/nightly/index.html)
* [Vagrant](https://vagrantup.com/) 1.6+ installed 

> NOTE: Ubuntu 14.04 LTS does not install a compatible version of Vagrant by default. You can upgrade the version that you get out of the box by downloading the `.deb` file from [Vagrant](http://www.vagrantup.com/downloads.html).

#####Spin up a virtual environment

Unzip the Lattice bundle, and switch to the vagrant directory

```bash
unzip lattice-bundle-VERSION-PLATFORM.zip
cd lattice-bundle-VERSION-PLATFORM/vagrant
vagrant up --provider virtualbox
```

This spins up a virtual environment that is accessible at `192.168.11.11`

#####Install the Lattice CLI

If you're running Linux: `curl -O http://receptor.192.168.11.11.xip.io/v1/sync/linux/ltc`

If you're running OS X: `curl -O http://receptor.192.168.11.11.xip.io/v1/sync/osx/ltc`

```bash
chmod a+x ltc
./ltc target 192.168.11.11.xip.io
./ltc -v
```

For more information visit [Lattice CLI](https://github.com/cloudfoundry-incubator/lattice/blob/master/ltc/README.md)

#####Use the Lattice CLI to target Lattice

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

#####Configure your virtual environment

Unzip the Lattice bundle, and switch to the terraform/aws directory

```bash
unzip lattice-bundle-VERSION-PLATFORM.zip
cd lattice-bundle-VERSION-PLATFORM/terraform/aws
```

Update the `lattice.aws.tf` by [filling in the values for the variables](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/aws/README.md#configure).

#####Deploy the cluster to AWS

From the folder containing the `lattice.aws.tf`, run these commands to deploy your cluster 

```bash
terraform get -update
terraform apply
```

Terraform will generate a `terraform.tfstate` file.  This file describes the cluster that was built - keep it around in order to modify/tear down the cluster.

#####Install the Lattice CLI

> NOTE: If your receptor has user/password authorization, you will need those credentials when downloading ltc for Terraform with curl.

If you're running Linux: `curl -Ou <username> http://receptor.192.168.11.11.xip.io/v1/sync/linux/ltc`

If you're running OS X: `curl -Ou <username> http://receptor.192.168.11.11.xip.io/v1/sync/osx/ltc`

```bash
chmod a+x ltc
./ltc target 192.168.11.11.xip.io
./ltc -v
```

For more information visit [Lattice CLI](https://github.com/cloudfoundry-incubator/lattice/blob/master/ltc/README.md)

#####Use the Lattice CLI to target Lattice

After a successful deployment Terraform will print the Lattice target and Lattice user information. Use the `lattice_target = x.x.x.x.xip.io` to target Lattice.

```bash
ltc target x.x.x.x.xip.io
```

[Lattice.cf](http://lattice.cf) | [Latest Release](https://github.com/cloudfoundry-incubator/lattice/releases/latest) | [Nightly Builds](https://lattice.s3.amazonaws.com/nightly/index.html)

