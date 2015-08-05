resource "openstack_compute_keypair_v2" "lattice-key" {
    name = "${var.openstack_key_name}"
    region = "${var.openstack_region}"
    public_key = "${var.openstack_public_key}"
}

resource "openstack_compute_secgroup_v2" "lattice-sg" {
    region = "${var.openstack_region}"
    name = "lattice-sg"
    description = "Security Group for Lattice"
    rule {
        from_port = 22
        to_port = 22
        ip_protocol = "tcp"
        cidr = "0.0.0.0/0"
    }
    rule {
        from_port = 1
        to_port = 65535
        ip_protocol = "tcp"
        cidr = "0.0.0.0/0"
    }
    rule {
        from_port = 1
        to_port = 65535
        ip_protocol = "udp"
        cidr = "0.0.0.0/0"
    }
    rule {
        from_port = 1
        to_port = 65535
        ip_protocol = "tcp"
        self = true
    }
    rule {
        from_port = 1
        to_port = 65535
        ip_protocol = "udp"
        self = true
    }
    rule {
        from_port = 1
        to_port = 1
        ip_protocol = "icmp"
        self = true
    }
}


resource "openstack_networking_network_v2" "lattice-network" {
    region = "${var.openstack_region}"
    name = "lattice-network"
    admin_state_up = "true"
}

resource "openstack_networking_subnet_v2" "lattice-network" {
    region = "${var.openstack_region}"
    network_id = "${openstack_networking_network_v2.lattice-network.id}"
    cidr = "${var.openstack_subnet_cidr_block}"
    ip_version = 4
}

resource "openstack_networking_router_v2" "lattice-network" {
    region = "${var.openstack_region}"
    name = "lattice-network"
    admin_state_up = "true"
    external_gateway = "${var.openstack_neutron_router_gateway_network_id}"
}

resource "openstack_networking_router_interface_v2" "lattice-network" {
    region = "${var.openstack_region}"
    router_id = "${openstack_networking_router_v2.lattice-network.id}"
    subnet_id = "${openstack_networking_subnet_v2.lattice-network.id}"
}

resource "openstack_compute_floatingip_v2" "fip-1" {
    region = "${var.openstack_region}"
    pool = "${var.openstack_floating_ip_pool_name}"
}

resource "openstack_compute_floatingip_v2" "fip-worker" {
    count = "${var.num_cells}"
    region = "${var.openstack_region}"
    pool = "${var.openstack_floating_ip_pool_name}"
}

resource "openstack_compute_instance_v2" "lattice-coordinator" {
    region = "${var.openstack_region}"
    name = "lattice-coordinator"
    image_name = "${var.openstack_image}"
    flavor_name = "${var.openstack_instance_type_coordinator}"
    key_pair = "${var.openstack_key_name}"
    security_groups = ["${openstack_compute_secgroup_v2.lattice-sg.name}"]
    metadata {
        lattice-role = "coordinator"
    }
    network {
        uuid = "${openstack_networking_network_v2.lattice-network.id}"
    }
    floating_ip = "${openstack_compute_floatingip_v2.fip-1.address}"

    connection {
        user = "${var.openstack_ssh_user}"
        key_file = "${var.openstack_ssh_private_key_file}"
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

    provisioner "remote-exec" {
        inline = [
            "sudo apt-get update",
            "sudo apt-get -y install btrfs-tools",
        ]
    }
    #/COMMON

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup",
            "sudo sh -c 'echo \"LATTICE_USERNAME=${var.lattice_username}\" > /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_PASSWORD=${var.lattice_password}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${openstack_compute_instance_v2.lattice-coordinator.access_ip_v4}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${openstack_compute_instance_v2.lattice-coordinator.floating_ip}.xip.io\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        inline = [
            "sudo apt-get -y install lighttpd lighttpd-mod-webdav",
            "sudo chmod 755 /tmp/install-from-tar",
            "sudo /tmp/install-from-tar brain",
        ]
    }
}

resource "openstack_compute_instance_v2" "lattice-cell" {
    count = "${var.num_cells}"
    region = "${var.openstack_region}"
    name = "lattice-cell-${count.index}"
    image_name = "${var.openstack_image}"
    flavor_name = "${var.openstack_instance_type_cell}"
    key_pair = "${var.openstack_key_name}"
    security_groups = ["${openstack_compute_secgroup_v2.lattice-sg.name}"]
    metadata {
        lattice-role = "cell"
        lattice-cell-instance = "${count.index}"
    }
    network {
        uuid = "${openstack_networking_network_v2.lattice-network.id}"
    }
    floating_ip = "${element(openstack_compute_floatingip_v2.fip-worker.*.address, count.index)}"

    connection {
        user = "${var.openstack_ssh_user}"
        key_file = "${var.openstack_ssh_private_key_file}"
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

    provisioner "remote-exec" {
        inline = [
            "sudo apt-get update",
            "sudo apt-get -y install btrfs-tools",
        ]
    }
    #/COMMON

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup",
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${openstack_compute_instance_v2.lattice-coordinator.access_ip_v4}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${openstack_compute_instance_v2.lattice-coordinator.floating_ip}.xip.io\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_CELL_ID=lattice-cell-${count.index}\" >> /var/lattice/setup/lattice-environment'",
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
