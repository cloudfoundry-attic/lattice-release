output "lattice_target" {
    value = "${aws_instance.lattice-brain.public_ip}.xip.io"
}

output "lattice_username" {
    value = "${var.lattice_username}"
}

output "lattice_password" {
    value = "${var.lattice_password}"
}
