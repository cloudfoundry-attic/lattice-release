##Running the box

    vagrant up
    
The box lives at 192.168.11.11
    
etcd is accessible on port 4001
Loggregator is accessible via loggregator.192.168.11.11.xip.io
receptor is accessible via receptor.192.168.11.11.xip.io

####Running Vagrant with a custom diego edge tar

    VAGRANT_DIEGO_EDGE_TAR_PATH=/vagrant/diego-edge.tgz vagrant up


##Testing the Diego Edge Box

 Follow the [whetstone instructions](https://github.com/pivotal-cf-experimental/whetstone) for diego-edge

##Working with Diego Edge

 Use the Diego Edge ClI: [Diego Edge Cli](https://github.com/pivotal-cf-experimental/diego-edge-cli).



##Developing
  Development work should be done on the develop branch.
  As a general rule, only CI should commit to master.
