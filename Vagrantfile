VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  system_ip = ENV["LATTICE_SYSTEM_IP"] || "192.168.11.11"
  config.vm.network "private_network", ip: system_ip

  config.vm.box = "lattice/ubuntu-trusty-64"
  config.vm.box_version = '0.2.0'

  config.vm.provision "shell" do |s|
    populate_lattice_env_file_script = <<-SCRIPT
      mkdir -pv /var/lattice/setup
      echo "CONSUL_SERVER_IP=#{system_ip}" >> /var/lattice/setup/lattice-environment
      echo "SYSTEM_DOMAIN=#{system_ip}.xip.io" >> /var/lattice/setup/lattice-environment
      echo "LATTICE_CELL_ID=lattice-cell-01" >> /var/lattice/setup/lattice-environment
      echo "GARDEN_EXTERNAL_IP=#{system_ip}" >> /var/lattice/setup/lattice-environment
    SCRIPT

    s.inline = populate_lattice_env_file_script
  end

  config.vm.provision "shell" do |s|
    s.inline = "cp /var/lattice/setup/lattice-environment /vagrant/.lattice-environment"
  end

  lattice_tar_version=File.read(File.join(File.dirname(__FILE__), "Version")).chomp
  system 'egrep -q \'\-[[:digit:]]+-g[0-9a-fA-F]{7,10}$\' ' + File.join(File.dirname(__FILE__), "Version")
  if $? == 0
    lattice_tar_url="https://s3-us-west-2.amazonaws.com/lattice/unstable/#{lattice_tar_version}/lattice.tgz"
  else
    lattice_tar_url="https://s3-us-west-2.amazonaws.com/lattice/releases/#{lattice_tar_version}/lattice.tgz"
  end

  config.vm.provision "shell" do |s|
    s.path = "install_from_tar"
    s.args = ["collocated", ENV["VAGRANT_LATTICE_TAR_PATH"].to_s, lattice_tar_url]
  end

  config.vm.provision "shell" do |s|
    s.inline = "export $(cat /var/lattice/setup/lattice-environment) && echo \"Lattice is now installed and running. You may target it with the Lattice cli via: ltc target $SYSTEM_DOMAIN\""
  end

end
