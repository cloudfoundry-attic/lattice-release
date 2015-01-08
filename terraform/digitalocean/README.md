# Lattice Terraform templates for DigitalOcean

This project contains [Terraform](https://www.terraform.io/) templates to help you deploy
[Lattice](https://github.com/pivotal-cf-experimental/lattice) on
[DigitalOcean](https://www.digitalocean.com).

## Usage

### Prerequisites

* A [DigitalOcean account](https://www.digitalocean.com)
* A [DigitalOcean API Token](https://www.digitalocean.com/community/tutorials/how-to-use-the-digitalocean-api-v2#how-to-generate-a-personal-access-token)
* A [DigitalOcean Password-less SSH Key](https://www.digitalocean.com/community/tutorials/how-to-use-ssh-keys-with-digitalocean-droplets)
* A DigitalOcean Region supporting [private networking](https://www.digitalocean.com/company/blog/introducing-private-networking/) (all regions except `sfo1`)

### Configure

Create a `lattice.tf` file (or use the provided [example](https://github.com/pivotal-cf-experimental/lattice/terraform/blob/master/digitalocean/lattice.tf.example)) and add the following contents updating the variables properly:

```
module "lattice" {
    source = "github.com/pivotal-cf-experimental/lattice/terraform/digitalocean"

    # Digital Ocean API token
    do_token = "<CHANGE-ME>"

    # SSH public key fingerprint
    do_ssh_public_key_fingerprint = "<CHANGE-ME>"

    # Path to the SSH private key file
    do_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch
    num_cells = "3"

    #################################
    ###  Optional Settings Below  ###
    #################################

    #If you wish to use your own lattice release instead of the latest version, uncomment the variable assignment below
    #and set it to your own lattice tar's path.
    # local_lattice_tar_path = ~/lattice.tgz

    # Digital Ocean region
    # do_region = "<CHANGE-ME>"
}
```

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

## Copyright

See [LICENSE](https://github.com/pivotal-cf-experimental/lattice/terraform/blob/master/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
