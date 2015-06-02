module "lattice-digitalocean" {
    source = "github.com/cloudfoundry-incubator/lattice//terraform//digitalocean?ref=v0.2.5"

    # Digital Ocean API token
    do_token = "<CHANGE-ME>"

    # SSH public key id. Get the key ID from https://developers.digitalocean.com/documentation/v1/ssh-keys/
    do_ssh_public_key_id = "<CHANGE-ME>"

    # Path to the SSH private key file. This needs to match the public key id defined above
    do_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch
    num_cells = "3"

    #################################
    ###  Optional Settings Below  ###
    #################################

    #If you wish to use your own lattice release instead of the latest version, uncomment the variable assignment below
    #and set it to your own lattice tar's path.
    # local_lattice_tar_path = "~/lattice.tgz"

    # Digital Ocean region
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
