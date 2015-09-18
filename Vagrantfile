Vagrant.configure("2") do |config|
  if Vagrant.has_plugin?("vagrant-proxyconf")
    config.proxy.http     = ENV["http_proxy"]
    config.proxy.https    = ENV["https_proxy"]
    config.proxy.no_proxy = [
      "localhost", "127.0.0.1",
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
    aws.instance_type = "m4.large"
    aws.ebs_optimized = true

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
  config.vm.box_version = '0.4.1'

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

  config.vm.provision "shell" do |s|
    s.inline = <<-SCRIPT
      tar xzf /vagrant/lattice.tgz --strip-components=2 -C /tmp lattice-build/scripts/install-from-tar
      /tmp/install-from-tar /vagrant/lattice.tgz
      . /var/lattice/setup/lattice-environment
      echo "Lattice is now installed and running."
      echo "You may target it using: ltc target ${SYSTEM_DOMAIN}\n"
    SCRIPT

    if provider_is_aws
      s.inline += <<-SCRIPT
        tar cf /var/lattice/garden/graph/warmrootfs.tar /var/lattice/rootfs
        rm -f /var/lattice/garden/graph/warmrootfs.tar
      SCRIPT
    end
  end
end

provision_required = (!ARGV.nil? && ['up', 'provision', 'reload'].include?(ARGV[0]))
lattice_tgz = File.join(File.dirname(__FILE__), "lattice.tgz")
lattice_url = defined?(LATTICE_URL) && LATTICE_URL

def download_lattice_tgz(url)
  unless system('curl', '-sfo', 'lattice.tgz', url)
    puts "Failed to download #{url}."
    exit(1)
  end
end

if provision_required && File.exists?(lattice_tgz)
  tgz_version = `tar Oxzf #{lattice_tgz} lattice-build/common/LATTICE_VERSION`.chomp
  if lattice_url
    url_version = lattice_url.match(/backend\/lattice-(v.+)\.tgz$/)[1]
    if tgz_version != url_version
      puts "Warning: lattice.tgz file version (#{tgz_version}) does not match Vagrantfile version (#{url_version})."
      puts 'Re-downloading and replacing local lattice.tgz...'
      download_lattice_tgz(lattice_url)
    end
  else
    repo_version = `git describe --tags --always`.chomp
    if tgz_version != repo_version && !ENV['IGNORE_VERSION_MISMATCH']
      puts '*******************************************************************************'
      puts "Error: lattice.tgz #{tgz_version} != current commit #{repo_version}\n"
      puts 'The lattice.tgz file was built using a different commit than the current one.'
      puts 'To ignore this error, set IGNORE_VERSION_MISMATCH=true in your environment.'
      puts "*******************************************************************************\n"
      exit(1)
    end
  end
end

if provision_required && !File.exists?(lattice_tgz)
  if lattice_url
    puts 'Local lattice.tgz not found, downloading...'
    download_lattice_tgz(lattice_url)
  else
    puts '*******************************************************************************'
    puts "Could not determine Lattice version, and no local lattice.tgz present.\n"
    puts 'As of v0.4.0, the process for deploying Lattice via Vagrant has changed.'
    puts 'Please use the process documented here:'
    puts "\thttp://github.com/cloudfoundry-incubator/lattice#launching-with-vagrant"
    puts "*******************************************************************************\n"
    exit(1)
  end
end
