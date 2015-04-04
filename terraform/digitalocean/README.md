# Lattice Terraform templates for DigitalOcean

This project contains [Terraform](https://www.terraform.io/) templates to help you deploy
[Lattice](https://github.com/cloudfoundry-incubator/lattice) on
[DigitalOcean](https://www.digitalocean.com).

## Usage

### Prerequisites

* A [DigitalOcean account](https://www.digitalocean.com)
* A [DigitalOcean API Token](https://www.digitalocean.com/community/tutorials/how-to-use-the-digitalocean-api-v2#how-to-generate-a-personal-access-token)
* A [DigitalOcean Password-less SSH Key](https://www.digitalocean.com/community/tutorials/how-to-use-ssh-keys-with-digitalocean-droplets)
* A DigitalOcean Region supporting [private networking](https://www.digitalocean.com/company/blog/introducing-private-networking/) (all regions except `sfo1`)

### Configure

Here are some step-by-step instructions for configuring a Lattice cluster via Terraform:

1. Visit the [Lattice GitHub Releases page](https://github.com/cloudfoundry-incubator/lattice/releases#)
2. Select the Lattice version you wish to deploy and download the Terraform example file for your target platform.  The filename will be `lattice.digitalocean.tf`
3. Create an empty folder and place the `lattice.digitalocean.tf` file in that folder.
4. Update the `lattice.digitalocean.tf` by filling in the values for the variables.  Details for the values of those variables are below.

The available variables that can be configured are:

* `do_token`: Digital Ocean API token
* `do_ssh_public_key_fingerprint`: SSH public key fingerprint
* `do_ssh_private_key_file`: Path to the SSH private key file
* `do_region`: The DO region to operate under (default `nyc2`)
* `do_image`: The droplet image ID or slug to base the launched instances (default `ubuntu-14-04-x64`)
* `do_size_coordinator`: The DO size to use for the Lattice Coordinator instance (default `512mb`)
* `do_size_cell`: The DO size to use for the Lattice Cell instances (default `2gb`)
* `num_cells`: The number of Lattice Cells to launch (default `3`)
* `lattice_username`: Lattice username (default `user`)
* `lattice_password`: Lattice password (default `pass`)

Refer to the [Terraform DigitalOcean (DO) provider](https://www.terraform.io/docs/providers/do/index.html)
documentation for more details about how to configure the proper credentials.

#### Generating the SSH public key fingerprint 

You can generate the SSH public key fingerprint from your public key via (e.g.)

```
ssh-keygen -lf ~/.ssh/id_rsa.pub
2048 aa:bb:cc:dd:ee:ff:aa:bb:cc:dd:ee:ff:aa:bb:cc:dd foo@bar.com (RSA)
```

The fingerprint is the second column in the output (`aa:bb...`)

### Deploy

Here are some step-by-step instructions for deploying a Lattice cluster via Terraform:

1. Run the following commands in the folder containing the `lattice.digitalocean.tf` file

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

### Use

Refer to the [Lattice CLI](https://github.com/cloudfoundry-incubator/lattice/tree/master/ltc) documentation.

### Destroy

Destroy the cluster:

```
terraform destroy
```

### Caveats

When working with Digital Ocean, terraform sometimes gets into an inconsistent state.  It is common for
the "terraform apply" command to time out while provisioning the VMs, and subsequent terraform commands will
then error out.  In some cases, a droplet will get stuck in the "New" status (not getting to "Active"), and
all Digital Ocean API commands will return:

```
* Error deleting droplet: Error destroying droplet: API Error: unprocessable_entity: Droplet already has a pending event.
```

In the event this happens, the recommended avenue is to use the Digital Ocean console to tear down all the droplets,
remove the terraform.tfstate file from the current directory, and then run "terraform apply" again to provision
from scratch.

## Copyright

See [LICENSE](https://github.com/cloudfoundry-incubator/lattice/blob/master/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
