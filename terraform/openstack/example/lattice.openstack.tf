module "lattice-openstack" {
    source = "github.com/cloudfoundry-incubator/lattice//terraform//openstack?ref=v0.3.1-15-g59bf81e"

    # Specify an API username and password for your lattice cluster
    # lattice_username = "<CHANGE-ME>"
    # lattice_password = "<CHANGE-ME>"

    # OpenStack User Account
    openstack_access_key = "<CHANGE-ME>"

    # OpenStack Password
    openstack_secret_key = "<CHANGE-ME>"

    # OpenStack Tenant Name
    openstack_tenant_name = "<CHANGE-ME>"

    # SSH Key Name
    openstack_key_name = "<CHANGE-ME>"

    # SSH Public Key to Upload
    openstack_public_key = "<CHANGE-ME>"

    # Path & filename of the SSH private key file
    openstack_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch
    num_cells = "3"

    # URI of Keystone authentication agent
    openstack_keystone_uri = "<CHANGE-ME>"

    # Instance Flavor Types
    openstack_instance_type_coordinator = "<CHANGE-ME>"
    openstack_instance_type_cell = "<CHANGE-ME>"

    # The internet-facing network which Neutron L3 routers should use as a gateway (UUID)
    openstack_neutron_router_gateway_network_id = "<CHANGE-ME>"

    # The name of the pool that floating IP addresses will be requested from
    openstack_floating_ip_pool_name = "<CHANGE-ME>"

    # The name of the Openstack Glance image used to spin up all VM instances.
    openstack_image = "<CHANGE-ME>"

    #################################
    ###  Optional Settings Below  ###
    #################################

    # If you wish to use your own lattice release instead of the latest version, 
    # uncomment the variable assignment below and set it to your own lattice tar's path.
    # local_lattice_tar_path = "~/lattice.tgz"

    # Openstack Region (Blank default for 'no region' installations)
    # openstack_region = "<CHANGE-ME>"
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
