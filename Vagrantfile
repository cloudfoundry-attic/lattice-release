Vagrant.configure("2") do |config|

  if Vagrant.has_plugin?("vagrant-proxyconf")
    config.proxy.http     = ENV["http_proxy"]
    config.proxy.https    = ENV["https_proxy"]
    config.proxy.no_proxy = [
      "localhost", 
      "127.0.0.1", 
      (ENV["LATTICE_SYSTEM_IP"] || "192.168.11.11"), 
      (ENV["LATTICE_SYSTEM_DOMAIN"] || "192.168.11.11.xip.io"), 
      ".consul"
    ].join(',')
  end

  ## credit: https://stefanwrobel.com/how-to-make-vagrant-performance-not-suck
  config.vm.provider "virtualbox" do |v|
    host = RbConfig::CONFIG['host_os']
    if host =~ /darwin/
      cpus = `sysctl -n hw.ncpu`.to_i
      mem = `sysctl -n hw.memsize`.to_i / 1024 / 1024 / 4
    elsif host =~ /linux/
      cpus = `nproc`.to_i
      mem = `grep 'MemTotal' /proc/meminfo | sed -e 's/MemTotal://' -e 's/ kB//'`.to_i / 1024 / 4
    else
      cpus = 2
      mem = 2048
    end

    v.customize ["modifyvm", :id, "--memory", mem]
    v.customize ["modifyvm", :id, "--cpus", cpus]
    v.customize ["modifyvm", :id, "--ioapic", "on"]
  end

  config.vm.provider :aws do |aws, override|
    aws.access_key_id = ENV["AWS_ACCESS_KEY_ID"]
    aws.secret_access_key = ENV["AWS_SECRET_ACCESS_KEY"]
    aws.keypair_name = "concourse-test"
    aws.instance_type = "m3.large"

    override.ssh.username = "ubuntu"
    override.ssh.private_key_path = ENV["AWS_SSH_PRIVATE_KEY_PATH"]
    override.nfs.functional = false
  end

  provider_is_aws = (!ARGV.nil? && ARGV.join(' ').match(/provider(=|\s+)aws/))

  if provider_is_aws
    system_values = <<-SCRIPT
      PUBLIC_IP=$(curl http://169.254.169.254/latest/meta-data/public-ipv4)
      PRIVATE_IP=$(hostname -I|awk '{print $1}')
      SYSTEM_DOMAIN="${PUBLIC_IP}.xip.io"
    SCRIPT

    config.ssh.insert_key = false
  else
    system_ip = ENV["LATTICE_SYSTEM_IP"] || "192.168.11.11"
    system_domain = ENV["LATTICE_SYSTEM_DOMAIN"] || "#{system_ip}.xip.io"

    system_values = <<-SCRIPT
      PUBLIC_IP=#{system_ip}
      PRIVATE_IP=#{system_ip}
      SYSTEM_DOMAIN=#{system_domain}
    SCRIPT

    config.vm.network "private_network", ip: system_ip
  end

  config.vm.box = "lattice/ubuntu-trusty-64"
  config.vm.box_version = '0.4.0'

  config.vm.provision "shell" do |s|
    s.inline = <<-SCRIPT
      mkdir -pv /var/lattice/setup

      #{system_values}
      echo "CONSUL_SERVER_IP=$PRIVATE_IP" >> /var/lattice/setup/lattice-environment
      echo "SYSTEM_IP=$PUBLIC_IP" >> /var/lattice/setup/lattice-environment
      echo "SYSTEM_DOMAIN=$SYSTEM_DOMAIN" >> /var/lattice/setup/lattice-environment
      echo "GARDEN_EXTERNAL_IP=$PRIVATE_IP" >> /var/lattice/setup/lattice-environment
      echo "LATTICE_CELL_ID=cell-01" >> /var/lattice/setup/lattice-environment
    SCRIPT
  end

  if Vagrant.has_plugin?("vagrant-proxyconf")
    config.vm.provision "shell" do |s|
      s.inline = "grep -i proxy /etc/environment >> /var/lattice/setup/lattice-environment || true"
    end
  end

  provision_required = (!ARGV.nil? && ['up', 'provision', 'reload'].include?(ARGV[0]))
  if provision_required && !File.exists?(File.join(File.dirname(__FILE__), "lattice.tgz"))
    lattice_url = defined?(LATTICE_URL) && LATTICE_URL

    if !lattice_url
      puts "Could not determine Lattice version, and no local lattice.tgz present.\n"
      puts '*******************************************************************************'
      puts 'As of v0.4.0, the process for deploying Lattice via Vagrant has changed.'
      puts 'Please use the process documented here: http://github.com/cloudfoundry-incubator/lattice#launching-with-vagrant'
      puts '*******************************************************************************'
      puts "\n"*3
      exit(1)
    end

    puts 'Local lattice.tgz not found, downloading...'
    unless system('curl', '-sfo', 'lattice.tgz', lattice_url)
      puts "Failed to download #{lattice_url}."
      exit(1)
    end
  end

  config.vm.provision "shell" do |s|
    s.inline = <<-SCRIPT
      tar xzf /vagrant/lattice.tgz --strip-components=2 -C /tmp lattice-build/scripts/install-from-tar
      /tmp/install-from-tar collocated /vagrant/lattice.tgz
      . /var/lattice/setup/lattice-environment
      echo "Lattice is now installed and running."
      echo "You may target it using: ltc target ${SYSTEM_DOMAIN}\n"
    SCRIPT
  end
end
