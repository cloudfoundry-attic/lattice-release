provider "openstack" {
     user_name = "${var.openstack_access_key}"
     tenant_name = "${var.openstack_tenant_name}"
     auth_url = "${var.openstack_keystone_uri}"
     password = "${var.openstack_secret_key}"
}
