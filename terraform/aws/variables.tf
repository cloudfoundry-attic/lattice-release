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
        ap-northeast-1 = "ami-4ceaed4d"
        ap-southeast-1 = "ami-f75875a5"
        ap-southeast-2 = "ami-950b62af"
        eu-central-1 = "ami-b01524ad"
        eu-west-1 = "ami-863686f1"
        sa-east-1 = "ami-cb70c1d6"
        us-east-1 = "ami-76e27e1e"
        us-west-1 = "ami-d5180890"
        us-west-2 = "ami-838dd9b3"
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
