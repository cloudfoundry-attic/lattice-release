VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  ## credit: https://stefanwrobel.com/how-to-make-vagrant-performance-not-suck
  config.vm.provider "virtualbox" do |v|
    host = RbConfig::CONFIG['host_os']

    # Give VM 1/4 system memory & access to all cpu cores on the host
    if host =~ /darwin/
      cpus = `sysctl -n hw.ncpu`.to_i
      # sysctl returns Bytes and we need to convert to MB
      mem = `sysctl -n hw.memsize`.to_i / 1024 / 1024 / 4
    elsif host =~ /linux/
      cpus = `nproc`.to_i
      # meminfo shows KB and we need to convert to MB
      mem = `grep 'MemTotal' /proc/meminfo | sed -e 's/MemTotal://' -e 's/ kB//'`.to_i / 1024 / 4
    else # sorry Windows folks, I can't help you
      cpus = 2
      mem = 2048
    end

    v.customize ["modifyvm", :id, "--memory", mem]
    v.customize ["modifyvm", :id, "--cpus", cpus]
    v.customize ["modifyvm", :id, "--ioapic", "on"]
  end

  system_ip = ENV["LATTICE_SYSTEM_IP"] || "192.168.11.11"
  system_domain = ENV["LATTICE_SYSTEM_DOMAIN"] || "#{system_ip}.xip.io"
  config.vm.network "private_network", ip: system_ip

  config.vm.box = "lattice/ubuntu-trusty-64"
  config.vm.box_version = '0.2.5'

  config.vm.provision "shell" do |s|
    populate_lattice_env_file_script = <<-SCRIPT
      mkdir -pv /var/lattice/setup
      echo "CONSUL_SERVER_IP=#{system_ip}" >> /var/lattice/setup/lattice-environment
      echo "SYSTEM_DOMAIN=#{system_domain}" >> /var/lattice/setup/lattice-environment
      echo "LATTICE_CELL_ID=cell-01" >> /var/lattice/setup/lattice-environment
      echo "GARDEN_EXTERNAL_IP=#{system_ip}" >> /var/lattice/setup/lattice-environment
    SCRIPT

    s.inline = populate_lattice_env_file_script
  end

  config.vm.provision "shell" do |s|
    s.inline = "cp /var/lattice/setup/lattice-environment /vagrant/.lattice-environment"
  end

  if Vagrant.has_plugin?("vagrant-proxyconf")
    config.vm.provision "shell" do |s|
      s.inline = "grep -i proxy /etc/environment >> /var/lattice/setup/lattice-environment || true"
    end
  end

  lattice_tar_version=File.read(File.join(File.dirname(__FILE__), "Version")).chomp
  if lattice_tar_version =~ /\-[[:digit:]]+\-g[0-9a-fA-F]{7,10}$/ 
    lattice_tar_url="https://s3-us-west-2.amazonaws.com/lattice/unstable/#{lattice_tar_version}/lattice.tgz"
  else
    lattice_tar_url="https://s3-us-west-2.amazonaws.com/lattice/releases/#{lattice_tar_version}/lattice.tgz"
  end

  config.vm.provision "shell" do |s|
    s.path = "cluster/scripts/install-from-tar"
    s.args = ["collocated", ENV["VAGRANT_LATTICE_TAR_PATH"].to_s, lattice_tar_url]
  end

  config.vm.provision "shell" do |s|
    s.inline = "export $(cat /var/lattice/setup/lattice-environment) && echo \"Lattice is now installed and running. You may target it with the Lattice cli via: ltc target $SYSTEM_DOMAIN\""
  end

end
