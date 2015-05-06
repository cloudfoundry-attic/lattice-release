variable "aws_access_key" {
    description = "AWS access key."
}

variable "aws_secret_key" {
    description = "AWS secret key."
}

variable "aws_region" {
    description = "AWS region."
    default = "us-east-1"
}

variable "aws_key_name" {
    description = "The SSH key name to use for the instances."
}

variable "aws_ssh_private_key_file" {
    description = "Path to the SSH private key file."
}

variable "aws_ssh_user" {
    description = "SSH user."
    default = "ubuntu"
}

variable "aws_vpc_cidr_block" {
    description = "The IPv4 address range that machines in the network are assigned to, represented as a CIDR block."
    default = "10.0.0.0/16"
}

variable "aws_subnet_cidr_block" {
    description = "The IPv4 address range that machines in the network are assigned to, represented as a CIDR block."
    default = "10.0.1.0/24"
}

variable "aws_image" {
    description = "The name of the image to base the launched instances."
    default = {
        ap-northeast-1 = "ami-66834c66"
        ap-southeast-1 = "ami-50ffc002"
        ap-southeast-2 = "ami-57136c6d"
        eu-central-1 = "ami-d6cef1cb"
        eu-west-1 = "ami-e3395194"
        sa-east-1 = "ami-8d8d0a90"
        us-east-1 = "ami-2411044c"
        us-west-1 = "ami-a7d738e3"
        us-west-2 = "ami-2bbb8a1b"
    }
}

variable "aws_instance_type_brain" {
    description = "The machine type to use for the Lattice Brain instance."
    default = "m3.medium"
}

variable "aws_instance_type_cell" {
    description = "The machine type to use for the Lattice Cells instances."
    default = "m3.medium"
}

variable "num_cells" {
    description = "The number of Lattice Cells to launch."
    default = "3"
}

variable "lattice_username" {
    description = "Lattice username."
    default = "user"
}

variable "lattice_password" {
    description = "Lattice password."
    default = "pass"
}

variable "local_lattice_tar_path" {
    description = "Path to the lattice tar, to deploy to your cluster. If not provided, then by default, the provisioner will download the latest lattice tar to a .lattice directory within your module path"
    default=".lattice/lattice.tgz"
}
