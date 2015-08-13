variable "do_token" {
    description = "Digital Ocean API token."
}

variable "do_ssh_public_key_id" {
    description = "SSH public key id."
}

variable "do_ssh_private_key_file" {
    description = "Path to the SSH private key file."
}

variable "do_region" {
    description = "The DO region to operate under."
    default = "nyc2"
}

variable "do_image" {
    description = "The droplet image ID or slug to base the launched instances."
    default = "ubuntu-14-04-x64"
}

variable "do_size_brain" {
    description = "The DO size to use for the Lattice Brain instance."
    default = "1gb"
}

variable "do_size_cell" {
    description = "The DO size to use for the Lattice Cell instances."
    default = "2gb"
}

variable "num_cells" {
    description = "The number of Lattice Cells to launch."
    default = "3"
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
