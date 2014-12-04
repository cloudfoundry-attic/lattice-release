##Running the box

    $ vagrant up

Vagrant up spins up a virtual hardware environment with an ip address assigned through DHCP. The output from vagrant up will provide instructions on how to target Diego. 

Use the [Diego Edge Cli](https://github.com/pivotal-cf-experimental/diego-edge-cli) to target Diego.

####Example Usage

     $ vagrant up
     
     Bringing machine 'default' up with 'vmware_fusion' provider...
     ...
     ...
     Diego-Edge is now installed and running. You may target it with the Diego-Edge cli via:
     192.168.194.130.xip.io
     
     $ diego-edge-cli target 192.168.194.130.xip.io 
     

####Running Vagrant with a custom diego edge tar

    VAGRANT_DIEGO_EDGE_TAR_PATH=/vagrant/diego-edge.tgz vagrant up


##Testing the Diego Edge Box

 Follow the [whetstone instructions](https://github.com/pivotal-cf-experimental/whetstone) for diego-edge

##Working with Diego Edge

 Use the Diego Edge ClI: [Diego Edge Cli](https://github.com/pivotal-cf-experimental/diego-edge-cli).



##Developing
  Development work should be done on the develop branch.
  As a general rule, only CI should commit to master.
