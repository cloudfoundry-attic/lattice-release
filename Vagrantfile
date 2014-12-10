VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.network "private_network", ip: "192.168.11.11"
  config.vm.box = "diego-edge/ubuntu-trusty-64"
  config.vm.box_version = '0.1.2'

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

end
