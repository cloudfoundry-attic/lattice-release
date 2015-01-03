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
        ap-northeast-1 = "ami-cec3cbcf"
        ap-southeast-1 = "ami-d16a4483"
        ap-southeast-2 = "ami-974c27ad"
        eu-central-1 = "ami-1a3b0b07"
        eu-west-1 = "ami-88df64ff"
        sa-east-1 = "ami-7d77c460"
        us-east-1 = "ami-d8a6c9b0"
        us-west-1 = "ami-71c8d534"
        us-west-2 = "ami-992d7ea9"
    }
}

variable "aws_instance_type_coordinator" {
    description = "The machine type to use for the Lattice Coordinator instance."
    default = "m3.medium"
}

variable "aws_instance_type_cell" {
    description = "The machine type to use for the Lattice Cells instances."
    default = "m3.xlarge"
}

variable "num_cells" {
    description = "The number of Lattice Cells to launch."
    default = "3"
}

variable "lattice_username" {
    description = "Lattice username."
    default = "admin"
}

variable "lattice_password" {
    description = "Lattice password."
    default = "c1oudc0w"
}
