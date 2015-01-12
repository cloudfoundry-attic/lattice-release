#Vagrant

##Running the box

    $ git clone git@github.com:pivotal-cf-experimental/lattice.git
    $ cd lattice
    $ vagrant up

Vagrant up spins up a virtual hardware environment that is accessible at 192.168.11.11. You can do this with either VMware Fusion or VirtualBox:. You can specify your preferred provider with the --provider flag.

Virtualbox:

     $ vagrant up --provider virtualbox

VMware Fusion:

     $ vagrant up --provider vmware_fusion

## Networking Conflicts
If you are trying to run both the Virtual Box and VMWare providers on the same machine, 
you'll need to run them on different private networks (subnets) that do not conflict.

Set the System IP to an address that does not conflict with the host networking configuration by passing the
LATTICE_SYSTEM_IP environment variable to the vagrant up command:

```
LATTICE_SYSTEM_IP=192.168.80.100 vagrant up
```

The output from vagrant up will provide instructions on how to target Lattice. 

Use the [Lattice Cli](https://github.com/pivotal-cf-experimental/lattice-cli) to target Lattice.

##Example Usage

     $ vagrant up
     
     Bringing machine 'default' up with 'vmware_fusion' provider...
     ...
     ...
     Lattice is now installed and running. You may target it with the Lattice cli via:
     192.168.194.130.xip.io
     
     $ ltc target 192.168.194.130.xip.io 
     

##Updating

Currently, Lattice does not support updating via provision.
So to update, you have to destroy the box and bring it back up as shown below:

     vagrant destroy --force
     git pull
     vagrant up
  
##Troubleshooting
-  xip.io is sometimes flaky, resulting in no such host errors.
-  The alternative that we have found is to use dnsmasq configured to resolve all xip.io addresses to 192.168.11.11.
-  This also requires creating a /etc/resolvers/io file that points to 127.0.0.1. See further instructions [here] (http://passingcuriosity.com/2013/dnsmasq-dev-osx/). 

##Running Vagrant with a custom lattice tar

    VAGRANT_LATTICE_TAR_PATH=/vagrant/lattice.tgz vagrant up

#Terraform Deployment


This project contains several [Terraform](https://www.terraform.io/) templates to help you deploy
[Lattice](https://github.com/pivotal-cf-experimental/lattice) on your choice of IaaS.
They are located under the terraform directory.

## Usage

### Prerequisites

* [Terraform](https://www.terraform.io/intro/getting-started/install.html) >= 0.3.6 installed on your machine
* Credentials for your choice of IaaS

### Configure

#### [Amazon Web Services](http://aws.amazon.com/):

Create a `lattice.tf` file by downloading the [AWS example file](https://github.com/pivotal-cf-experimental/lattice/blob/master/terraform/aws/lattice.tf.example):

``` bash
wget --quiet https://raw.githubusercontent.com/pivotal-cf-experimental/lattice/master/terraform/aws/lattice.tf.example -O lattice.tf
```

Update the downloaded file filling the variables according to the [AWS README](https://github.com/pivotal-cf-experimental/lattice/blob/master/terraform/aws/README.md) file.

#### [DigitalOcean](https://www.digitalocean.com):

Create a `lattice.tf` file by downloading the [DigitalOcean example file](https://github.com/pivotal-cf-experimental/lattice/blob/master/terraform/digitalocean/lattice.tf.example):

``` bash
wget --quiet https://raw.githubusercontent.com/pivotal-cf-experimental/lattice/master/terraform/digitalocean/lattice.tf.example -O lattice.tf
```

Update the downloaded file filling the variables according to the [DigitalOcean README](https://github.com/pivotal-cf-experimental/lattice/blob/master/terraform/digitalocean/README.md) file.

#### [Google Cloud](https://cloud.google.com/):

Create a `lattice.tf` file downloading the [Google Cloud example file](https://github.com/pivotal-cf-experimental/lattice/blob/master/terraform/google/lattice.tf.example):

``` bash
wget --quiet https://raw.githubusercontent.com/pivotal-cf-experimental/lattice/lattice-terraform/master/google/lattice.tf.example -O lattice.tf
```
Update the downloaded file filling the variables according to the [Google Cloud README](https://github.com/pivotal-cf-experimental/lattice/blob/master/terraform/google/README.md) file.

### Deploy

Get the templates and deploy the cluster:

```
terraform get -update
terraform apply
```

After the cluster has been successfully, terraform will print the Lattice domain:

```
Outputs:

  lattice_target = x.x.x.x.xip.io
  lattice_username = xxxxxxxx
  lattice_password = xxxxxxxx
```



## Destroy

Destroy the cluster:

```
terraform destroy
```



#Testing the Lattice Box

 Follow the [whetstone instructions](https://github.com/pivotal-cf-experimental/whetstone) for lattice

#Using Lattice

 Use the [Lattice Cli](https://github.com/pivotal-cf-experimental/lattice-cli).



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
* by closing [issues](https://github.com/pivotal-cf-experimental/lattice/issues)
* by reviewing patches

##Development Workflow

Development work should be done on the develop branch.
As a general rule, only CI should commit to master.

## Submitting an Issue
We use the [GitHub issue tracker](https://github.com/pivotal-cf-experimental/lattice/issues) to track bugs and features.
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

See [LICENSE](https://github.com/pivotal-cf-experimental/lattice/blob/master/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
