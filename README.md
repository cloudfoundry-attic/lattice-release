##Running the box

    $ vagrant up

Vagrant up spins up a virtual hardware environment that is accessible at 192.168.11.11. You can do this with either VMware Fusion or VirtualBox:. You can specify your preferred provider with the --provider flag.

Virtualbox:

     $ vagrant up --provider virtualbox

VMware Fusion:

     $ vagrant up --provider vmware_fusion

### Networking Conflicts
If you are trying to run both the Virtual Box and VMWare providers on the same machine, 
you'll need to run them on different private networks that do not conflict. 

Change the line of the vagrant file that says

      config.vm.network "private_network", ip: "192.168.11.11"

to 

      config.vm.network "private_network", ip: "192.168.80.100"

or a different non-conflicting IP.

The output from vagrant up will provide instructions on how to target Diego. 

Use the [Lattice Cli](https://github.com/pivotal-cf-experimental/lattice-cli) to target Diego.

####Example Usage

     $ vagrant up
     
     Bringing machine 'default' up with 'vmware_fusion' provider...
     ...
     ...
     Diego-Edge is now installed and running. You may target it with the Diego-Edge cli via:
     192.168.194.130.xip.io
     
     $ ltc target 192.168.194.130.xip.io 
     

##Updating

Currently, Diego Edge does not support updating via provision.
So to update, you have to destroy the box and bring it back up as shown below:

  vagrant destroy --force
  git pull
  vagrant up

####Running Vagrant with a custom diego edge tar

    VAGRANT_DIEGO_EDGE_TAR_PATH=/vagrant/diego-edge.tgz vagrant up


##Testing the Diego Edge Box

 Follow the [whetstone instructions](https://github.com/pivotal-cf-experimental/whetstone) for diego-edge

##Working with Diego Edge

 Use the [Lattice Cli](https://github.com/pivotal-cf-experimental/lattice-cli).


##Developing
  Development work should be done on the develop branch.
  As a general rule, only CI should commit to master.
