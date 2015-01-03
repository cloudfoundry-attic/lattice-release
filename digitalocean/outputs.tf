output "lattice_target" {
    value = "${digitalocean_droplet.lattice-coordinator.ipv4_address}.xip.io"
}

output "lattice_username" {
    value = "${var.lattice_username}"
}

output "lattice_password" {
    value = "${var.lattice_password}"
}
