resource "google_compute_network" "lattice" {
    name = "lattice"
    ipv4_range = "${var.gce_ipv4_range}"
}

resource "google_compute_firewall" "lattice" {
    name = "lattice"
    network = "${google_compute_network.lattice.name}"
    source_ranges = ["0.0.0.0/0"]
    allow {
        protocol = "tcp"
        ports = ["1-65535"]
    }
    allow {
        protocol = "udp"
        ports = ["1-65535"]
    }
    target_tags = ["lattice"]
}

resource "google_compute_address" "lattice-coordinator" {
    name = "lattice-coordinator"
}

resource "google_compute_instance" "lattice-coordinator" {
    zone = "${var.gce_zone}"
    name = "lattice-coordinator"
    tags = ["lattice"]
    description = "Lattice Coordinator"
    machine_type = "${var.gce_machine_type_coordinator}"
    disk {
        image = "${var.gce_image}"
        auto_delete = true
    }
    network {
        source = "${google_compute_network.lattice.name}"
        address = "${google_compute_address.lattice-coordinator.address}"
    }

    connection {
        user = "${var.gce_ssh_user}"
        key_file = "${var.gce_ssh_private_key_file}"
    }

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup/",
            "sudo sh -c 'echo \"LATTICE_USERNAME=${var.lattice_username}\" > /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_PASSWORD=${var.lattice_password}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_COORDINATOR_IP=${google_compute_address.lattice-coordinator.address}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${google_compute_address.lattice-coordinator.address}.xip.io\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        scripts = [
            "${path.module}/../scripts/install_lattice_common",
            "${path.module}/../scripts/install_lattice_coordinator",
        ]
    }
}

resource "google_compute_instance" "lattice-cell" {
    count = "${var.num_cells}"
    zone  = "${var.gce_zone}"
    name  = "lattice-cell-${count.index}"
    tags  = ["lattice"]
    description = "Lattice Cell ${count.index}"
    machine_type = "${var.gce_machine_type_cell}"
    disk {
        image = "${var.gce_image}"
        auto_delete = true
    }
    network {
        source = "${google_compute_network.lattice.name}"
    }

    connection {
        user = "${var.gce_ssh_user}"
        key_file = "${var.gce_ssh_private_key_file}"
    }

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup/",
            "sudo sh -c 'echo \"LATTICE_USERNAME=${var.lattice_username}\" > /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_PASSWORD=${var.lattice_password}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_COORDINATOR_IP=${google_compute_address.lattice-coordinator.address}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${google_compute_address.lattice-coordinator.address}.xip.io\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"DIEGO_CELL_ID=lattice-cell-${count.index}\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        scripts = [
            "${path.module}/../scripts/install_lattice_common",
            "${path.module}/../scripts/install_lattice_cell",
        ]
    }
}
