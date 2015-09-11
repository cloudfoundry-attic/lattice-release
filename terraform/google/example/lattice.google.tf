module "lattice-google" {
    # Specify a source containing the terraform configuration
    # source = "<CHANGE ME>"

    # Specify a URL or local path to a lattice.tgz file for deployment
    # lattice_tar_source = "<CHANGE-ME>"

    # Specify an API username and password for your lattice cluster
    # lattice_username = "<CHANGE-ME>"
    # lattice_password = "<CHANGE-ME>"

    # Path to the JSON file used to describe your account credentials, downloaded from Google Cloud Console
    # gce_account_file = "<CHANGE-ME>"

    # The name of the project to apply any resources to
    # gce_project = "<CHANGE-ME>"

    # SSH user
    # gce_ssh_user = "<CHANGE-ME>"

    # Path to the SSH private key file
    # gce_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch (optional, default: 1)
    # num_cells = "1"

    # Google Compute Engine zone (optional, default: "us-central1-a")
    # gce_zone = "us-central1-a"

    # Namespace (optional, default: "lattice")
    # lattice_namespace = "lattice"
}

output "lattice_target" {
    value = "${module.lattice-google.lattice_target}"
}

output "lattice_username" {
    value = "${module.lattice-google.lattice_username}"
}

output "lattice_password" {
    value = "${module.lattice-google.lattice_password}"
}
