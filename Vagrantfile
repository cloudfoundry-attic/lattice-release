VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.network "private_network", type: "dhcp"

  config.vm.provision "shell" do |s|
    s.path = "install_from_tar"
    s.args = ENV["VAGRANT_DIEGO_EDGE_TAR_PATH"]
  end

  config.vm.provision "shell" do |s|
    s.inline = "cp /var/diego/system-domain /vagrant/.system-domain"
  end

  config.vm.provision "shell" do |s|
    s.inline = "echo 'Diego-Edge is now installed and running. You may target it with the Diego-Edge cli via:' && cat /vagrant/.system-domain"
  end

  config.vm.provider "virtualbox" do |v, override|
    # dns resolution appears to be very slow in some environments; this fixes it
    v.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]

    # increase memory on provisioned vm to 4gb
    v.customize ["modifyvm", :id, "--memory", 4096]
  end

  config.vm.provider "vmware_fusion" do |v, override|
    override.vm.box = "casualjim/trusty-vagrant"
    # increase memory on provisioned vm to 4gb
    v.vmx["memsize"] = "4096"
  end
end
