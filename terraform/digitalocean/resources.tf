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
            "sudo apt-get -y upgrade",
            "sudo apt-get -y install curl",
            "sudo apt-get -y install gcc",
            "sudo apt-get -y install make",
            "sudo apt-get -y install quota",
            "sudo apt-get -y install linux-image-extra-$(uname -r)",
            "sudo apt-get -y install btrfs-tools",
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
        inline = [
            "sudo apt-get -y install lighttpd lighttpd-mod-webdav",
            "sudo chmod 755 /tmp/install-from-tar",
            "sudo /tmp/install-from-tar brain",
        ]
    }
}

resource "digitalocean_droplet" "cell" {
    count    = "${var.num_cells}"
    name     = "cell-${count.index}"
    region   = "${var.do_region}"
    image    = "${var.do_image}"
    size     = "${var.do_size_cell}"
    ssh_keys = [
        "${var.do_ssh_public_key_id}",
    ]
    private_networking = true

    connection {
        key_file = "${var.do_ssh_private_key_file}"
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
            "sudo apt-get -y upgrade",
            "sudo apt-get -y install curl",
            "sudo apt-get -y install gcc",
            "sudo apt-get -y install make",
            "sudo apt-get -y install quota",
            "sudo apt-get -y install linux-image-extra-$(uname -r)",
            "sudo apt-get -y install btrfs-tools",
        ]
    }
    #/COMMON

    provisioner "remote-exec" {
        inline = [
            "sudo mkdir -p /var/lattice/setup/",
            "sudo sh -c 'echo \"CONSUL_SERVER_IP=${digitalocean_droplet.lattice-brain.ipv4_address}\" >> /var/lattice/setup/lattice-environment'",
            "sudo sh -c 'echo \"SYSTEM_DOMAIN=${digitalocean_droplet.lattice-brain.ipv4_address}.xip.io\" >> /var/lattice/setup/lattice-environment'",
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
