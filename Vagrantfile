VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  system_ip = "192.168.11.11"
  config.vm.network "private_network", ip: system_ip

  config.vm.box = "lattice/ubuntu-trusty-64"
  config.vm.box_version = '0.1.3'

  config.vm.provision "shell" do |s|
    s.path = "install_from_tar"
    s.args = ["collocated", ENV["VAGRANT_LATTICE_TAR_PATH"].to_s]
  end

  config.vm.provision "shell" do |s|
    s.inline = "cp /var/lattice/setup/lattice-environment /vagrant/.lattice-environment"
  end

  config.vm.provision "shell" do |s|
    s.inline = "export $(cat /var/lattice/setup/lattice-environment) && echo \"Lattice is now installed and running. You may target it with the Lattice cli via: ltc target $SYSTEM_DOMAIN\""
  end

  config.vm.provision "shell" do |s|
    populate_lattice_env_file_script = <<-SCRIPT
      echo "CONSUL_SERVER_IP=#{system_ip}" >> /var/lattice/setup/lattice-environment
      echo "SYSTEM_DOMAIN=#{system_ip}.xip.io" >> /var/lattice/setup/lattice-environment
      echo "DIEGO_CELL_ID=lattice-cell-01" >> /var/lattice/setup/lattice-environment
    SCRIPT

    s.inline = populate_lattice_env_file_script
  end

end
