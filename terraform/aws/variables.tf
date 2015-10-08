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
        ap-northeast-1 = "ami-720b9d72"
        ap-southeast-1 = "ami-d4adb886"
        ap-southeast-2 = "ami-43317f79"
        eu-central-1 = "ami-864e4d9b"
        eu-west-1 = "ami-430f2334"
        sa-east-1 = "ami-312fba2c"
        us-east-1 = "ami-935a2bf6"
        us-west-1 = "ami-7775b033"
        us-west-2 = "ami-f7514fc7"
    }
}

variable "aws_instance_type_brain" {
    description = "The machine type to use for the Lattice Brain instance."
    default = "m4.large"
}

variable "aws_instance_type_cell" {
    description = "The machine type to use for the Lattice Cells instances."
    default = "m4.large"
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

variable "lattice_tar_source" {
    description = "URL or local path of the lattice tar used to deploy your cluster."
}
