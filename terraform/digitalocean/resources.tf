resource "digitalocean_droplet" "lattice-coordinator" {
    name     = "lattice-coordinator"
    region   = "${var.do_region}"
    image    = "${var.do_image}"
    size     = "${var.do_size_coordinator}"
    ssh_keys = [
      "${var.do_ssh_public_key_fingerprint}",
    ]
    private_networking = true

    connection {
        key_file = "${var.do_ssh_private_key_file}"
    }

    #COMMON
    provisioner "local-exec" {
      command = "LOCAL_LATTICE_TAR_PATH=${var.local_lattice_tar_path} ${path.module}/../local-scripts/download-lattice-tar"
    }

    provisioner "file" {
      source = "${var.local_lattice_tar_path}"
      destination = "/tmp/lattice.tgz"
    }

    provisioner "file" {
      source = "${path.module}/../remote-scripts/install_from_tar"
      destination = "/tmp/install_from_tar"
    }

    provisioner "remote-exec" {
      inline = [
          "sudo chmod 755 /tmp/install_from_tar",
          "sudo bash -c \"echo 'PATH_TO_LATTICE_TAR=${var.local_lattice_tar_path}' >> /etc/environment\""
      ]
    }
    #/COMMON

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup/",
            "sudo sh -c 'echo \"LATTICE_USERNAME=${var.lattice_username}\" > /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_PASSWORD=${var.lattice_password}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${digitalocean_droplet.lattice-coordinator.ipv4_address}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${digitalocean_droplet.lattice-coordinator.ipv4_address}.xip.io\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        script = "${path.module}/../remote-scripts/install-lattice-coordinator"
    }
}

resource "digitalocean_droplet" "lattice-cell" {
    count    = "${var.num_cells}"
    name     = "lattice-cell-${count.index}"
    region   = "${var.do_region}"
    image    = "${var.do_image}"
    size     = "${var.do_size_cell}"
    ssh_keys = [
      "${var.do_ssh_public_key_fingerprint}",
    ]
    private_networking = true

    connection {
        key_file = "${var.do_ssh_private_key_file}"
    }

    #COMMON
    provisioner "local-exec" {
      command = "LOCAL_LATTICE_TAR_PATH=${var.local_lattice_tar_path} ${path.module}/../local-scripts/download-lattice-tar"
    }

    provisioner "file" {
      source = "${var.local_lattice_tar_path}"
      destination = "/tmp/lattice.tgz"
    }

    provisioner "file" {
      source = "${path.module}/../remote-scripts/install_from_tar"
      destination = "/tmp/install_from_tar"
    }

    provisioner "remote-exec" {
      inline = [
          "sudo chmod 755 /tmp/install_from_tar",
          "sudo bash -c \"echo 'PATH_TO_LATTICE_TAR=${var.local_lattice_tar_path}' >> /etc/environment\""
      ]
    }
    #/COMMON

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup/",
            "sudo sh -c 'echo \"LATTICE_USERNAME=${var.lattice_username}\" > /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_PASSWORD=${var.lattice_password}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${digitalocean_droplet.lattice-coordinator.ipv4_address}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${digitalocean_droplet.lattice-coordinator.ipv4_address}.xip.io\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"DIEGO_CELL_ID=lattice-cell-${count.index}\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        script = "${path.module}/../remote-scripts/install-lattice-cell"
    }
}
