# Lattice Terraform templates for Openstack

This project contains [Terraform](https://www.terraform.io/) templates to help you deploy
[Lattice](https://github.com/cloudfoundry-incubator/lattice) on
[Openstack](http://www.openstack.org/). 

> Note: This is a community-supplied platform, it is not presently supported by the project maintainers.

## Usage

### Prerequisites

* An [Openstack Cloud](http://www.openstack.org/)
* An Account, project and relevant access on your openstack cloud.

### Configure

Here are some step-by-step instructions for configuring a Lattice cluster via Terraform:

1. Visit the [Lattice GitHub Releases page](https://github.com/cloudfoundry-incubator/lattice/releases)
2. Select the Lattice version you wish to deploy and download the Terraform example file for your target platform.  The filename will be `lattice.openstack.tf`
3. Create an empty folder and place the `lattice.openstack.tf` file in that folder.
4. Update the `lattice.openstack.tf` by filling in the values for the variables.  Details for the values of those variables are below.

The available variables that can be configured are:

* `openstack_access_key`: Openstack username.
* `openstack_secret_key`: Openstack Password.
* `openstack_tenant_name`: The Tenant/Project name in Openstack.
* `openstack_key_name`: The name given to the SSH key which will be uploaded for use by the instances.
* `openstack_public_key`: The actual contents of rsa_id.pub to upload as the public key.
* `openstack_ssh_private_key_file`: Path to the SSH private key file (Stays local. Used for provisioning.)
* `openstack_ssh_user`: SSH user (default `ubuntu`)
* `num_cells`: The number of Lattice Cells to launch (default `3`)
* `openstack_keystone_uri`: The Keystone API URL
*  `openstack_instance_type_coordinator`: flavour for Coordinator node
* `openstack_instance_type_cell`: flavour for Cell nodes
* `openstack_neutron_router_gateway_network_id`: The UUID of a network which can be used by a Neutron router as a WAN IP address. (Gateway Network)
* `openstack_floating_ip_pool_name`: The Name of the Openstack IP pool from which to request floating IP's
* `openstack_image`: The Ubuntu image in glance to use for Instances.
* `openstack_region`: Openstack Region. Leave blank to support Openstack installations without region support.
* `lattice_username`: Lattice username (default `user`)
* `lattice_password`: Lattice password (default `pass`)


Refer to the [Terraform AWS provider](https://www.terraform.io/docs/providers/openstack/index.html)
documentation for more details about how to configure the proper credentials.

### Deploy

Here are some step-by-step instructions for deploying a Lattice cluster via Terraform:

1. Run the following commands in the folder containing the `lattice.openstack.tf` file

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

Sometimes, destroy will need to be run twice to completely destroy all components, as openstack networking components are often seen as 'still in use' when destroyed immidietley after the instances that relied on them.


## Updating

The provided examples (i.e., `lattice.openstack.tf`) are pinned to a specific Bump commit or release tag in order to maintain compatibility between the Lattice build (`lattice.tgz`) and the Terraform definitions.  Currently, Terraform does not automatically update to newer revisions of Lattice.  

If you want to update to the latest version of Lattice:  
  - Update the `ref` in the `source` directive of your `lattice.openstack.tf` to `master`.
  - Run `terraform get -update` to update the modules under the `.terraform/` folder.
 
If you want to update to a specific version of Lattice:
  - Choose a version from either the [Bump commits](https://github.com/cloudfoundry-incubator/lattice/commits/master) or [Releases](https://github.com/cloudfoundry-incubator/lattice/releases).
  - Update the `ref` in the `source` directive of your `lattice.openstack.tf` to that version.
  - Run `terraform get -update` to update the modules under the `.terraform/` folder.


## Copyright
Openstack Terraform support for Lattice.cf added by Matt Johnson <matjohn2@cisco.com>.

For Lattice copyright, see [LICENSE](../../docs/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
