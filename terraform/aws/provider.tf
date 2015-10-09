provider "aws" {
    access_key = "${var.aws_access_key_id}"
    secret_key = "${var.aws_secret_access_key}"
    region = "${var.aws_region}"
}

output "target" {
    value = "${aws_eip.ip.public_ip}.xip.io"
}

output "username" {
    value = "${var.username}"
}

output "password" {
    value = "${var.password}"
}
