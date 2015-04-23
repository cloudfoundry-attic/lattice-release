resource "google_compute_network" "lattice-network" {
    name = "lattice"
    ipv4_range = "${var.gce_ipv4_range}"
}

resource "google_compute_firewall" "lattice-network" {
    name = "lattice"
    network = "${google_compute_network.lattice-network.name}"
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

resource "google_compute_address" "lattice-brain" {
    name = "lattice-brain"
}

resource "google_compute_instance" "lattice-brain" {
    zone = "${var.gce_zone}"
    name = "lattice-brain"
    tags = ["lattice"]
    description = "Lattice Brain"
    machine_type = "${var.gce_machine_type_brain}"
    disk {
        image = "${var.gce_image}"
        auto_delete = true
    }
    network {
        source = "${google_compute_network.lattice-network.name}"
        address = "${google_compute_address.lattice-brain.address}"
    }

    connection {
        user = "${var.gce_ssh_user}"
        key_file = "${var.gce_ssh_private_key_file}"
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
          "sudo chmod 755 /tmp/install-from-tar",
          "sudo bash -c \"echo 'PATH_TO_LATTICE_TAR=${var.local_lattice_tar_path}' >> /etc/environment\""
      ]
    }
    #/COMMON

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup/",
            "sudo sh -c 'echo \"LATTICE_USERNAME=${var.lattice_username}\" > /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_PASSWORD=${var.lattice_password}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${google_compute_address.lattice-brain.address}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${google_compute_address.lattice-brain.address}.xip.io\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        script = "${path.module}/../scripts/remote/install-brain"
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
        source = "${google_compute_network.lattice-network.name}"
    }

    connection {
        user = "${var.gce_ssh_user}"
        key_file = "${var.gce_ssh_private_key_file}"
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
          "sudo chmod 755 /tmp/install-from-tar",
          "sudo bash -c \"echo 'PATH_TO_LATTICE_TAR=${var.local_lattice_tar_path}' >> /etc/environment\""
      ]
    }
    #/COMMON

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup/",
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${google_compute_address.lattice-brain.address}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${google_compute_address.lattice-brain.address}.xip.io\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_CELL_ID=lattice-cell-${count.index}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"GARDEN_EXTERNAL_IP=$(hostname -I | awk '\"'\"'{ print $1 }'\"'\"')\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        script = "${path.module}/../scripts/remote/install-cell"
    }
}
