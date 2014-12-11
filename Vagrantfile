VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.network "private_network", ip: "192.168.11.11"
  config.vm.box = "diego-edge/ubuntu-trusty-64" #TODO: CHANGE TO LATTICE AND UPLOAD BOXES
  config.vm.box_version = '0.1.3'

  config.vm.provision "shell" do |s|
    s.path = "install_from_tar"
    s.args = ENV["VAGRANT_LATTICE_TAR_PATH"]
  end

  config.vm.provision "shell" do |s|
    s.inline = "cp /var/lattice/setup/lattice-environment /vagrant/.lattice-environment"
  end

  config.vm.provision "shell" do |s|
    s.inline = "export $(cat /var/lattice/setup/lattice-environment) && echo \"Lattice is now installed and running. You may target it with the Lattice cli via: $SYSTEM_DOMAIN\""
  end

end
