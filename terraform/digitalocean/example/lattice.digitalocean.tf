module "lattice-digitalocean" {
    # Specify a source containing the terraform configuration
    # source = "<CHANGE ME>"

    # Specify a URL or local path to a lattice.tgz file for deployment
    # lattice_tar_source = "<CHANGE-ME>"

    # Specify an API username and password for your lattice cluster
    # lattice_username = "<CHANGE-ME>"
    # lattice_password = "<CHANGE-ME>"

    # Digital Ocean API token
    do_token = "<CHANGE-ME>"

    # SSH public key id. Get the key ID from https://developers.digitalocean.com/documentation/v1/ssh-keys/
    do_ssh_public_key_id = "<CHANGE-ME>"

    # Path to the SSH private key file. This needs to match the public key id defined above
    do_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch
    num_cells = "3"

    # Digital Ocean region (optional)
    # do_region = "<CHANGE-ME>"
}

output "lattice_target" {
    value = "${module.lattice-digitalocean.lattice_target}"
}

output "lattice_username" {
    value = "${module.lattice-digitalocean.lattice_username}"
}

output "lattice_password" {
    value = "${module.lattice-digitalocean.lattice_password}"
}
