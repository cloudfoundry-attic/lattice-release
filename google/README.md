# Lattice Terraform templates for Google Cloud

This project contains [Terraform](https://www.terraform.io/) templates to help you deploy
[Lattice](https://github.com/pivotal-cf-experimental/lattice) on
[Google Cloud](https://cloud.google.com/).

## Usage

### Prerequisites

* A [Google Cloud account](https://cloud.google.com/)
* A [Google Compute Engine project](https://cloud.google.com/compute/docs/projects)
* A [Google Compute Engine account and client secrets files](https://www.terraform.io/docs/providers/google/index.html)
* A [Google Compute Engine Password-less SSH Key](https://cloud.google.com/compute/docs/console#sshkeys)
* [Terraform](https://www.terraform.io/downloads.html) >= 0.3.6 (not yet released as January 2nd 2015, [![hashicorp/terraform/pull/724](https://github-shields.cfapps.io/github/hashicorp/terraform/pull/724.svg)](https://github-shields.cfapps.io/github/hashicorp/terraform/pull/724) should be merged)

### Configure

Create a `lattice.tf` file (or use the provided [example](https://github.com/cf-platform-eng/lattice-terraform/blob/master/google/lattice.tf.example)) and add the following contents updating the variables properly:

```
module "lattice" {
    source = "github.com/cf-platform-eng/lattice-terraform/google"

    # Path to the JSON file used to describe your account credentials, downloaded from Google Cloud Console
    gce_account_file = "<CHANGE-ME>"

    # Path to the JSON file containing the secrets for your account, downloaded from Google Cloud Console
    gce_client_secrets_file = "<CHANGE-ME>"

    # The name of the project to apply any resources to
    gce_project = "<CHANGE-ME>"

    # SSH user
    gce_ssh_user = "<CHANGE-ME>"

    # Path to the SSH private key file
    gce_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch
    num_cells = 3
}
```

The available variables that can be configured are:

* `gce_account_file`: Path to the JSON file used to describe your account credentials, downloaded from Google Cloud Console
* `gce_client_secrets_file`: Path to the JSON file containing the secrets for your account, downloaded from Google Cloud Console
* `gce_project`: The name of the project to apply any resources to
* `gce_ssh_user`: SSH user
* `gce_ssh_private_key_file`: Path to the SSH private key file
* `gce_region`: The region to operate under (default `us-central1`)
* `gce_zone`: The zone that the machines should be created in (default `us-central1-a`)
* `gce_ipv4_range`: The IPv4 address range that machines in the network are assigned to, represented as a CIDR block (default `10.0.0.0/16`)
* `gce_image`: The name of the image to base the launched instances (default `ubuntu-1404-trusty-v20141212`)
* `gce_machine_type_coordinator`: The machine type to use for the Lattice Coordinator instance (default `n1-standard-1`)
* `gce_machine_type_cell`: The machine type to use for the Lattice Cells instances (default `n1-standard-4`)
* `num_cells`: The number of Lattice Cells to launch (default `3`)
* `lattice_username`: Lattice username (default `admin`)
* `lattice_password`: Lattice password (default `c1oudc0w`)

Refer to the [Terraform Google Cloud provider](https://www.terraform.io/docs/providers/google/index.html)
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

See [LICENSE](https://github.com/cf-platform-eng/lattice-terraform/blob/master/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
