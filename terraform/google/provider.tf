provider "google" {
    account_file = "${var.gce_account_file}"
    client_secrets_file = "${var.gce_client_secrets_file}"
    project = "${var.gce_project}"
    region = "${var.gce_region}"
}
