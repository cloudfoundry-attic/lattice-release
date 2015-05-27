# Lattice Terraform templates for Google Cloud

This project contains [Terraform](https://www.terraform.io/) templates to help you deploy
[Lattice](https://github.com/cloudfoundry-incubator/lattice) on
[Google Cloud](https://cloud.google.com/).

## Usage

### Prerequisites

* A [Google Cloud account](https://cloud.google.com/)
* A [Google Compute Engine project](https://cloud.google.com/compute/docs/projects)
* A [Google Compute Engine account file](https://www.terraform.io/docs/providers/google/index.html)
* A [Google Compute Engine Password-less SSH Key](https://cloud.google.com/compute/docs/console#sshkeys)

### Configure

Here are some step-by-step instructions for configuring a Lattice cluster via Terraform:

1. Visit the [Lattice GitHub Releases page](https://github.com/cloudfoundry-incubator/lattice/releases)
2. Select the Lattice version you wish to deploy and download the Terraform example file for your target platform.  The filename will be `lattice.google.tf`
3. Create an empty folder and place the `lattice.google.tf` file in that folder.
4. Update the `lattice.google.tf` by filling in the values for the variables.  Details for the values of those variables are below.

The available variables that can be configured are:

* `gce_account_file`: Path to the JSON file used to describe your account credentials, downloaded from Google Cloud Console
* `gce_project`: The name of the project to apply any resources to
* `gce_ssh_user`: SSH user
* `gce_ssh_private_key_file`: Path to the SSH private key file
* `gce_region`: The region to operate under (default `us-central1`)
* `gce_zone`: The zone that the machines should be created in (default `us-central1-a`)
* `gce_ipv4_range`: The IPv4 address range that machines in the network are assigned to, represented as a CIDR block (default `10.0.0.0/16`)
* `gce_image`: The name of the image to base the launched instances (default `ubuntu-1404-trusty-v20141212`)
* `gce_machine_type_brain`: The machine type to use for the Lattice Brain instance (default `n1-standard-1`)
* `gce_machine_type_cell`: The machine type to use for the Lattice Cells instances (default `n1-standard-4`)
* `num_cells`: The number of Lattice Cells to launch (default `3`)
* `lattice_username`: Lattice username (default `user`)
* `lattice_password`: Lattice password (default `pass`)

Refer to the [Terraform Google Cloud provider](https://www.terraform.io/docs/providers/google/index.html)
documentation for more details about how to configure the proper credentials.

### Deploy

Here are some step-by-step instructions for deploying a Lattice cluster via Terraform:

1. Run the following commands in the folder containing the `lattice.google.tf` file

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

Refer to the [Lattice CLI](../../ltc) documentation.

### Destroy

Destroy the cluster:

```
terraform destroy
```

## Updating

The provided examples (i.e., `lattice.google.tf`) are pinned to a specific Bump commit or release tag in order to maintain compatibility between the Lattice build (`lattice.tgz`) and the Terraform definitions.  Currently, Terraform does not automatically update to newer revisions of Lattice.  

If you want to update to the latest version of Lattice:  
  - Update the `ref` in the `source` directive of your `lattice.google.tf` to `master`.
  - Run `terraform get -update` to update the modules under the `.terraform/` folder.
 
If you want to update to a specific version of Lattice:
  - Choose a version from either the [Bump commits](https://github.com/cloudfoundry-incubator/lattice/commits/master) or [Releases](https://github.com/cloudfoundry-incubator/lattice/releases).
  - Update the `ref` in the `source` directive of your `lattice.google.tf` to that version.
  - Run `terraform get -update` to update the modules under the `.terraform/` folder.

## Copyright

See [LICENSE](../../docs/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
