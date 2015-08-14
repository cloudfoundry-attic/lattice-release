module "lattice-aws" {
    source = "github.com/cloudfoundry-incubator/lattice//terraform//aws?ref=v0.3.1-15-g59bf81e"

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

    #################################
    ###  Optional Settings Below  ###
    #################################

    # If you wish to use your own lattice release instead of the latest version, 
    # uncomment the variable assignment below and set it to your own lattice tar's path.
    # local_lattice_tar_path = "~/lattice.tgz"

    # AWS region (e.g., us-west-1)
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
