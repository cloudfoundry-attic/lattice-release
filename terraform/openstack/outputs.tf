output "lattice_target" {
    value = "${openstack_compute_instance_v2.lattice-coordinator.floating_ip}.xip.io"
}

output "lattice_username" {
    value = "${var.lattice_username}"
}

output "lattice_password" {
    value = "${var.lattice_password}"
}
