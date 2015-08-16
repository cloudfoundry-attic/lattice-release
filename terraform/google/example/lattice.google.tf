module "lattice-google" {
    source = "github.com/cloudfoundry-incubator/lattice//terraform//google?ref=v0.3.1-15-g59bf81e"

    # Specify an API username and password for your lattice cluster
    # lattice_username = "<CHANGE-ME>"
    # lattice_password = "<CHANGE-ME>"

    # Path to the JSON file used to describe your account credentials, downloaded from Google Cloud Console
    gce_account_file = "<CHANGE-ME>"

    # The name of the project to apply any resources to
    gce_project = "<CHANGE-ME>"

    # SSH user
    gce_ssh_user = "<CHANGE-ME>"

    # Path to the SSH private key file
    gce_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch
    num_cells = "1"

    #################################
    ###  Optional Settings Below  ###
    #################################

    # If you wish to use your own lattice release instead of the latest version, 
    # uncomment the variable assignment below and set it to your own lattice tar's path.
    # local_lattice_tar_path = "~/lattice.tgz"

    # Google Compute Engine zone
    # gce_zone = "<CHANGE-ME>"
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
