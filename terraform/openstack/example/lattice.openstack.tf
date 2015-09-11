module "lattice-openstack" {
    # Specify a source containing the terraform configuration
    # source = "<CHANGE ME>"

    # Specify a URL or local path to a lattice.tgz file for deployment
    # lattice_tar_source = "<CHANGE-ME>"

    # Specify an API username and password for your lattice cluster
    # lattice_username = "<CHANGE-ME>"
    # lattice_password = "<CHANGE-ME>"

    # OpenStack User Account
    # openstack_access_key = "<CHANGE-ME>"

    # OpenStack Password
    # openstack_secret_key = "<CHANGE-ME>"

    # OpenStack Tenant Name
    # openstack_tenant_name = "<CHANGE-ME>"

    # SSH Key Name
    # openstack_key_name = "<CHANGE-ME>"

    # SSH Public Key to Upload
    # openstack_public_key = "<CHANGE-ME>"

    # Path & filename of the SSH private key file
    # openstack_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch (optional, default: "3")
    # num_cells = "3"

    # URI of Keystone authentication agent
    # openstack_keystone_uri = "<CHANGE-ME>"

    # Instance Flavor Types (optional, defaults: "m3.medium")
    # openstack_instance_type_coordinator = "m3.medium"
    # openstack_instance_type_cell = "m3.medium"

    # The internet-facing network which Neutron L3 routers should use as a gateway (UUID)
    # openstack_neutron_router_gateway_network_id = "<CHANGE-ME>"

    # The name of the pool that floating IP addresses will be requested from
    # openstack_floating_ip_pool_name = "<CHANGE-ME>"

    # The name of the Openstack Glance image used to spin up all VM instances.
    # openstack_image = "<CHANGE-ME>"

    # Openstack Region (optional, default: "" [no region])
    # openstack_region = ""
}

output "lattice_target" {
    value = "${module.lattice-openstack.lattice_target}"
}

output "lattice_username" {
    value = "${module.lattice-openstack.lattice_username}"
}

output "lattice_password" {
    value = "${module.lattice-openstack.lattice_password}"
}
