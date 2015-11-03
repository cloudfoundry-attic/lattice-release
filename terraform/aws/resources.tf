resource "aws_vpc" "network" {
    cidr_block = "${var.vpc_cidr_block}"
    enable_dns_support = true
    enable_dns_hostnames = true

    tags {
        Name = "lattice"
    }
}

resource "aws_subnet" "network" {
    vpc_id = "${aws_vpc.network.id}"
    cidr_block = "${var.subnet_cidr_block}"
    map_public_ip_on_launch = true

    tags {
        Name = "lattice"
    }
}

resource "aws_internet_gateway" "network" {
    vpc_id = "${aws_vpc.network.id}"
}

resource "aws_route_table" "network" {
    vpc_id = "${aws_vpc.network.id}"

    route {
        cidr_block = "0.0.0.0/0"
        gateway_id = "${aws_internet_gateway.network.id}"
    }
}

resource "aws_route_table_association" "network" {
    subnet_id = "${aws_subnet.network.id}"
    route_table_id = "${aws_route_table.network.id}"
}

resource "aws_security_group" "network" {
    name = "lattice"
    description = "Lattice security group"
    vpc_id = "${aws_vpc.network.id}"

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

resource "aws_instance" "brain" {
    ami = "${lookup(var.brain_ami, var.aws_region)}"
    instance_type = "${var.brain_instance_type}"
    key_name = "${var.aws_ssh_private_key_name}"
    subnet_id = "${aws_subnet.network.id}"
    security_groups = ["${aws_security_group.network.id}"]

    tags {
        Name = "lattice-brain"
    }

    connection {
        user = "${var.ssh_username}"
        key_file = "${var.aws_ssh_private_key_path}"
    }

    provisioner "remote-exec" {
        inline = ["mkdir -p /tmp/terraform"]
    }

    provisioner "file" {
        source = "."
        destination = "/tmp/terraform"
    }
}

resource "aws_eip" "ip" {
    instance = "${aws_instance.brain.id}"
    vpc = true

    connection {
        host = "${aws_eip.ip.public_ip}"
        user = "${var.ssh_username}"
        key_file = "${var.aws_ssh_private_key_path}"
    }

    provisioner "remote-exec" {
        inline = [
            "sudo sh -c 'echo USERNAME=${var.username} >> /var/lattice/setup'",
            "sudo sh -c 'echo PASSWORD=${var.password} >> /var/lattice/setup'",
            "sudo sh -c 'echo DOMAIN=${aws_eip.ip.public_ip}.xip.io >> /var/lattice/setup'",
            "sudo sh -c 'echo HOST_ID=lattice-brain-0 >> /var/lattice/setup'",
            "mkdir -p /tmp/terraform/.lattice",
            "[ -f /tmp/terraform/.lattice/lattice.tgz ] || curl -s -o /tmp/terraform/.lattice/lattice.tgz '${var.lattice_tgz_url}'",
            "sudo tar xzf /tmp/terraform/.lattice/lattice.tgz -C /tmp install",
            "sudo /tmp/install/common",
            "sudo /tmp/install/brain /tmp/terraform/.lattice/lattice.tgz",
            "sudo /tmp/install/start",
            "rm -rf /tmp/terraform"
        ]
    }
}

resource "aws_instance" "cell" {
    depends_on = ["aws_eip.ip"]
    count = "${var.cell_count}"
    ami = "${lookup(var.cell_ami, var.aws_region)}"
    instance_type = "${var.cell_instance_type}"
    ebs_optimized = true
    key_name = "${var.aws_ssh_private_key_name}"
    subnet_id = "${aws_subnet.network.id}"
    security_groups = ["${aws_security_group.network.id}"]

    tags {
        Name = "lattice-cell-${count.index}"
    }

    connection {
        user = "${var.ssh_username}"
        key_file = "${var.aws_ssh_private_key_path}"
    }

    provisioner "remote-exec" {
        inline = ["mkdir -p /tmp/terraform"]
    }

    provisioner "file" {
        source = "."
        destination = "/tmp/terraform"
    }

    provisioner "remote-exec" {
        inline = [
            "sudo sh -c 'echo HOST_ID=lattice-cell-${count.index} >> /var/lattice/setup'",
            "sudo sh -c \"echo GARDEN_IP=$(ip route get 1 | awk '{print $NF;exit}') >> /var/lattice/setup\"",
            "sudo sh -c 'echo BRAIN_IP=${aws_instance.brain.private_ip} >> /var/lattice/setup'",
            "mkdir -p /tmp/terraform/.lattice",
            "[ -f /tmp/terraform/.lattice/lattice.tgz ] || curl -s -o /tmp/terraform/.lattice/lattice.tgz '${var.lattice_tgz_url}'",
            "sudo tar xzf /tmp/terraform/.lattice/lattice.tgz -C /tmp install",
            "sudo /tmp/install/common",
            "sudo /tmp/install/cell /tmp/terraform/.lattice/lattice.tgz",
            "sudo /tmp/install/terraform/cell",
            "sudo /tmp/install/start",
            "rm -rf /tmp/terraform"
        ]
    }
}
