module "lattice-aws" {
    # Specify a source containing the terraform configuration
    # source = "<CHANGE ME>"

    # Specify a URL or local path to a lattice.tgz file for deployment
    # lattice_tar_source = "<CHANGE-ME>"

    # Specify an API username and password for your lattice cluster
    # lattice_username = "<CHANGE-ME>"
    # lattice_password = "<CHANGE-ME>"

    # AWS access key
    aws_access_key = "<CHANGE-ME>"

    # AWS secret key
    aws_secret_key = "<CHANGE-ME>"

    # The SSH key name to use for the instances
    aws_key_name = "<CHANGE-ME>"

    # Path to the SSH private key file
    aws_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch
    num_cells = "3"

    # AWS region (optional)
    # aws_region = "<CHANGE-ME>"
}

output "lattice_target" {
    value = "${module.lattice-aws.lattice_target}"
}

output "lattice_username" {
    value = "${module.lattice-aws.lattice_username}"
}

output "lattice_password" {
    value = "${module.lattice-aws.lattice_password}"
}
