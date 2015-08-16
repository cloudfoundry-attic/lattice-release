variable "gce_account_file" {
    description = "Path to the JSON file used to describe your account credentials, downloaded from Google Cloud Console."
}

variable "gce_project" {
    description = "The name of the project to apply any resources to."
}

variable "gce_ssh_user" {
    description = "SSH user."
}

variable "gce_ssh_private_key_file" {
    description = "Path to the SSH private key file."
}

variable "gce_region" {
    description = "The region to operate under."
    default = "us-central1"
}

variable "gce_zone" {
    description = "The zone that the machines should be created in."
    default = "us-central1-a"
}

variable "gce_ipv4_range" {
    description = "The IPv4 address range that machines in the network are assigned to, represented as a CIDR block."
    default = "10.0.0.0/16"
}

variable "gce_image" {
    description = "The name of the image to base the launched instances."
    default = "ubuntu-1404-trusty-v20150316"
}

variable "gce_machine_type_brain" {
    description = "The machine type to use for the Lattice Brain instance."
    default = "n1-standard-1"
}

variable "gce_machine_type_cell" {
    description = "The machine type to use for the Lattice Cells instances."
    default = "n1-standard-4"
}

variable "num_cells" {
    description = "The number of Lattice Cells to launch."
    default = "1"
}

variable "lattice_username" {
    description = "Lattice username."
}

variable "lattice_password" {
    description = "Lattice password."
}

variable "local_lattice_tar_path" {
    description = "Path to the lattice tar, to deploy to your cluster. If not provided, then by default, the provisioner will download the latest lattice tar to a .lattice directory within your module path"
    default=".lattice/lattice.tgz"
}
