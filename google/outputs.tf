output "lattice_target" {
    value = "${google_compute_address.lattice-coordinator.address}.xip.io"
}

output "lattice_username" {
    value = "${var.lattice_username}"
}

output "lattice_password" {
    value = "${var.lattice_password}"
}
