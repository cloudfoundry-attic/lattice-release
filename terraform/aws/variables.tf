variable "lattice_tgz_url" {
    description = "URL for lattice.tgz"
}

variable "username" {
    description = "Lattice username"
}

variable "password" {
    description = "Lattice password"
}

variable "aws_access_key_id" {
    description = "AWS access key ID"
}

variable "aws_secret_access_key" {
    description = "AWS secret access key"
}

variable "aws_ssh_private_key_name" {
    description = "AWS SSH private key name"
}

variable "aws_ssh_private_key_path" {
    description = "AWS SSH private key file path"
}

variable "aws_region" {
    description = "AWS region"
    default = "us-east-1"
}

variable "cell_count" {
    description = "Number of Lattice cells to launch"
    default = "3"
}

variable "brain_instance_type" {
    description = "Machine type for the Lattice brain"
    default = "t2.medium"
}

variable "cell_instance_type" {
    description = "Machine type for Lattice cells"
    default = "m4.large"
}

variable "ssh_username" {
    description = "SSH username for base image AMI"
    default = "ubuntu"
}

variable "vpc_cidr_block" {
    description = "CIDR address range for AWS VPC"
    default = "10.0.0.0/16"
}

variable "subnet_cidr_block" {
    description = "CIDR address range for AWS subnet"
    default = "10.0.1.0/24"
}

variable "brain_ami" {
    description = "Brain base image AMI"
    default = {
        ap-northeast-1 = ""
        ap-southeast-1 = ""
        ap-southeast-2 = ""
        eu-central-1 = ""
        eu-west-1 = ""
        sa-east-1 = ""
        us-east-1 = "ami-8f80ceea"
        us-west-1 = ""
        us-west-2 = ""
    }
}

variable "cell_ami" {
    description = "Cell base image AMI"
    default = {
        ap-northeast-1 = ""
        ap-southeast-1 = ""
        ap-southeast-2 = ""
        eu-central-1 = ""
        eu-west-1 = ""
        sa-east-1 = ""
        us-east-1 = "ami-d580ceb0"
        us-west-1 = ""
        us-west-2 = ""
    }
}

