resource "aws_vpc" "lattice-network" {
    cidr_block = "${var.aws_vpc_cidr_block}"
    enable_dns_support = true
    enable_dns_hostnames = true
    tags {
        Name = "lattice"
    }
}

resource "aws_subnet" "lattice-network" {
    vpc_id = "${aws_vpc.lattice-network.id}"
    cidr_block = "${var.aws_subnet_cidr_block}"
    map_public_ip_on_launch = true
    tags {
        Name = "lattice"
    }
}

resource "aws_internet_gateway" "lattice-network" {
    vpc_id = "${aws_vpc.lattice-network.id}"
}

resource "aws_route_table" "lattice-network" {
    vpc_id = "${aws_vpc.lattice-network.id}"
    route {
        cidr_block = "0.0.0.0/0"
        gateway_id = "${aws_internet_gateway.lattice-network.id}"
    }
}

resource "aws_route_table_association" "lattice-network" {
    subnet_id = "${aws_subnet.lattice-network.id}"
    route_table_id = "${aws_route_table.lattice-network.id}"
}

resource "aws_security_group" "lattice-network" {
    name = "lattice"
    description = "lattice security group"
    vpc_id = "${aws_vpc.lattice-network.id}"
    ingress {
        protocol = "tcp"
        from_port = 1
        to_port = 65535
        cidr_blocks = ["0.0.0.0/0"]
    }
    ingress {
        protocol = "udp"
        from_port = 1
        to_port = 65535
        cidr_blocks = ["0.0.0.0/0"]
    }
    egress {
        protocol = "tcp"
        from_port = 1
        to_port = 65535
        cidr_blocks = ["0.0.0.0/0"]
    }
    egress {
        protocol = "udp"
        from_port = 1
        to_port = 65535
        cidr_blocks = ["0.0.0.0/0"]
    }
    tags {
        Name = "lattice"
    }
}

resource "aws_eip" "ip" {
    instance = "${aws_instance.lattice-brain.id}"
    vpc = true
    connection {
        host = "${aws_eip.ip.public_ip}"
        user = "${var.aws_ssh_user}"
        key_file = "${var.aws_ssh_private_key_file}"
        agent = false
    }
    provisioner "remote-exec" {
        inline = [       
          "sudo sh -c 'echo \"SYSTEM_DOMAIN=${aws_eip.ip.public_ip}.xip.io\" >> /var/lattice/setup/lattice-environment'",
          "sudo restart receptor",
          "sudo restart trafficcontroller"
        ]   
    }
}

resource "aws_instance" "lattice-brain" {
    ami = "${lookup(var.aws_image, var.aws_region)}"
    instance_type = "${var.aws_instance_type_brain}"
    key_name = "${var.aws_key_name}"
    subnet_id = "${aws_subnet.lattice-network.id}"
    security_groups = [
        "${aws_security_group.lattice-network.id}",
    ]
    tags {
        Name = "lattice-brain"
    }

    connection {
        user = "${var.aws_ssh_user}"
        key_file = "${var.aws_ssh_private_key_file}"
        agent = false
    }

    #COMMON
    provisioner "local-exec" {
        command = "LOCAL_LATTICE_TAR_PATH=${var.local_lattice_tar_path} LATTICE_VERSION_FILE_PATH=${path.module}/../../Version ${path.module}/../scripts/local/download-lattice-tar"
    }

    provisioner "file" {
        source = "${var.local_lattice_tar_path}"
        destination = "/tmp/lattice.tgz"
    }

    provisioner "file" {
        source = "${path.module}/../scripts/remote/install-from-tar"
        destination = "/tmp/install-from-tar"
    }
    #/COMMON

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup",
            "sudo sh -c 'echo \"LATTICE_USERNAME=${var.lattice_username}\" > /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_PASSWORD=${var.lattice_password}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${aws_instance.lattice-brain.private_ip}\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        inline = [
            "sudo chmod 755 /tmp/install-from-tar",
            "sudo /tmp/install-from-tar brain",
        ]
    }
}

resource "aws_instance" "cell" {
    depends_on = ["aws_eip.ip"]
    count = "${var.num_cells}"
    ami = "${lookup(var.aws_image, var.aws_region)}"
    instance_type = "${var.aws_instance_type_cell}"
    key_name = "${var.aws_key_name}"
    subnet_id = "${aws_subnet.lattice-network.id}"
    security_groups = [
        "${aws_security_group.lattice-network.id}",
    ]
    tags {
        Name = "cell-${count.index}"
    }

    connection {
        user = "${var.aws_ssh_user}"
        key_file = "${var.aws_ssh_private_key_file}"
        agent = false
    }

    #COMMON
    provisioner "local-exec" {
        command = "LOCAL_LATTICE_TAR_PATH=${var.local_lattice_tar_path} LATTICE_VERSION_FILE_PATH=${path.module}/../../Version ${path.module}/../scripts/local/download-lattice-tar"
    }

    provisioner "file" {
        source = "${var.local_lattice_tar_path}"
        destination = "/tmp/lattice.tgz"
    }

    provisioner "file" {
        source = "${path.module}/../scripts/remote/install-from-tar"
        destination = "/tmp/install-from-tar"
    }
    #/COMMON

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup",
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${aws_instance.lattice-brain.private_ip}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${aws_eip.ip.public_ip}.xip.io\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_CELL_ID=cell-${count.index}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"GARDEN_EXTERNAL_IP=$(hostname -I | awk '\"'\"'{ print $1 }'\"'\"')\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        inline = [
            "sudo chmod 755 /tmp/install-from-tar",
            "sudo /tmp/install-from-tar cell",
        ]
    }
}


