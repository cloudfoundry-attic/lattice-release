module "lattice-aws" {
    # Specify a source containing the terraform configuration
    # source = "<CHANGE ME>"

    # Specify a URL or local path to a lattice.tgz file for deployment
    # lattice_tar_source = "<CHANGE-ME>"

    # Specify an API username and password
    # username = "<CHANGE-ME>"
    # password = "<CHANGE-ME>"

    # The number of Lattice Cells to launch (optional, default: "3")
    # cell_count = "3"

    # AWS access key
    # aws_access_key = "<CHANGE-ME>"

    # AWS secret key
    # aws_secret_key = "<CHANGE-ME>"

    # The SSH key name to use for the instances
    # aws_ssh_private_key_name = "<CHANGE-ME>"

    # Path to the SSH private key file
    # aws_ssh_private_key_path = "<CHANGE-ME>"

    # AWS region (optional, default: "us-east-1")
    # aws_region = "us-east-1"
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
