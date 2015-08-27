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
        ap-northeast-1 = "ami-8453ec84"
        ap-southeast-1 = "ami-f0aca0a2"
        ap-southeast-2 = "ami-63c58559"
        eu-central-1 = "ami-08f5f115"
        eu-west-1 = "ami-16d08161"
        sa-east-1 = "ami-a1840dbc"
	us-east-1 = "ami-19dd6672"
	us-west-1 = "ami-1108f055"
	us-west-2 = "ami-d5ddcce5"
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
}

variable "lattice_password" {
    description = "Lattice password."
}

variable "lattice_tar_source" {
    description = "URL or local path of the lattice tar used to deploy your cluster."
}
