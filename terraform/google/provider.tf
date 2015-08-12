provider "google" {
    account_file = "${file("${var.gce_account_file}")}"
    project = "${var.gce_project}"
    region = "${var.gce_region}"
}
