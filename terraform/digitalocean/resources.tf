resource "digitalocean_droplet" "lattice-brain" {
    name     = "lattice-brain"
    region   = "${var.do_region}"
    image    = "${var.do_image}"
    size     = "${var.do_size_brain}"
    ssh_keys = [
      "${var.do_ssh_public_key_id}",
    ]
    private_networking = true

    connection {
        key_file = "${var.do_ssh_private_key_file}"
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
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${digitalocean_droplet.lattice-brain.ipv4_address}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${digitalocean_droplet.lattice-brain.ipv4_address}.xip.io\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        script = "${path.module}/../scripts/remote/install-brain"
    }
}

resource "digitalocean_droplet" "lattice-cell" {
    count    = "${var.num_cells}"
    name     = "lattice-cell-${count.index}"
    region   = "${var.do_region}"
    image    = "${var.do_image}"
    size     = "${var.do_size_cell}"
    ssh_keys = [
      "${var.do_ssh_public_key_id}",
    ]
    private_networking = true

    connection {
        key_file = "${var.do_ssh_private_key_file}"
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
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${digitalocean_droplet.lattice-brain.ipv4_address}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${digitalocean_droplet.lattice-brain.ipv4_address}.xip.io\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"LATTICE_CELL_ID=lattice-cell-${count.index}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"GARDEN_EXTERNAL_IP=$(hostname -I | awk '\"'\"'{ print $1 }'\"'\"')\" >> /var/lattice/setup/lattice-environment'",
        ]
    }

    provisioner "remote-exec" {
        script = "${path.module}/../scripts/remote/install-cell"
    }
}
