# Lattice Terraform templates

This project contains several [Terraform](https://www.terraform.io/) templates to help you deploy
[Lattice](https://github.com/pivotal-cf-experimental/lattice) on your choice of IaaS.

## Usage

### Prerequisites

* [Terraform](https://www.terraform.io/intro/getting-started/install.html) installed on your machine
* Credentials for your choice of IaaS

### Configure

#### [Amazon Web Services](http://aws.amazon.com/):

Create a `lattice.tf` file downloading the [AWS example file](https://github.com/cf-platform-eng/lattice-terraform/blob/master/aws/lattice.tf.example):

``` bash
wget --quiet https://raw.githubusercontent.com/cf-platform-eng/lattice-terraform/master/aws/lattice.tf.example -O lattice.tf
```

Update the downloaded file filling the variables according to the [AWS README](https://github.com/cf-platform-eng/lattice-terraform/blob/master/aws/README.md) file.

#### [DigitalOcean](https://www.digitalocean.com):

Create a `lattice.tf` file downloading the [DigitalOcean example file](https://github.com/cf-platform-eng/lattice-terraform/blob/master/digitalocean/lattice.tf.example):

``` bash
wget --quiet https://raw.githubusercontent.com/cf-platform-eng/lattice-terraform/master/digitalocean/lattice.tf.example -O lattice.tf
```

Update the downloaded file filling the variables according to the [DigitalOcean README](https://github.com/cf-platform-eng/lattice-terraform/blob/master/digitalocean/README.md) file.

#### [Google Cloud](https://cloud.google.com/):

Create a `lattice.tf` file downloading the [Google Cloud example file](https://github.com/cf-platform-eng/lattice-terraform/blob/master/google/lattice.tf.example):

``` bash
wget --quiet https://raw.githubusercontent.com/cf-platform-eng/lattice-terraform/master/google/lattice.tf.example -O lattice.tf
```
Update the downloaded file filling the variables according to the [Google Cloud README](https://github.com/cf-platform-eng/lattice-terraform/blob/master/google/README.md) file.

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

### Use

Refer to the [Lattice CLI](https://github.com/pivotal-cf-experimental/lattice-cli#lattice-cli) documentation.

### Destroy

Destroy the cluster:

```
terraform destroy
```

## Contributing

In the spirit of [free software](http://www.fsf.org/licensing/essays/free-sw.html), **everyone** is encouraged to help improve this project.

Here are some ways *you* can contribute:

* by using alpha, beta, and prerelease versions
* by reporting bugs
* by suggesting new features
* by writing or editing documentation
* by writing specifications
* by writing code (**no patch is too small**: fix typos, add comments, clean up inconsistent whitespace)
* by refactoring code
* by closing [issues](https://github.com/cf-platform-eng/lattice-terraform/issues)
* by reviewing patches


### Submitting an Issue
We use the [GitHub issue tracker](https://github.com/cf-platform-eng/lattice-terraform/issues) to track bugs and features.
Before submitting a bug report or feature request, check to make sure it hasn't already been submitted.
You can indicate support for an existing issue by voting it up.
When submitting a bug report, please include a [Gist](http://gist.github.com/) that includes a stack trace and any
details that may be necessary to reproduce the bug, including your gem version, Ruby version, and operating system.
Ideally, a bug report should include a pull request with failing specs.

### Submitting a Pull Request

1. Fork the project.
2. Create a topic branch.
3. Implement your feature or bug fix.
4. Commit and push your changes.
5. Submit a pull request.

## Copyright

See [LICENSE](https://github.com/cf-platform-eng/lattice-terraform/blob/master/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
