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
 
    
##To provision a box from scratch:
    
Clone Diego-Edge and its dependencies cf-release and diego-edge and compile the necessary binaries (Currently this is run outside of the guest vm)
   
    mkdir -p ~/workspace
    cd workspace
    git clone git@github.com:pivotal-cf-experimental/diego-edge.git
    git clone git@github.com:cloudfoundry-incubator/diego-release.git
    git clone git@github.com:cloudfoundry/cf-release.git
   
    cd diego-edge
    scripts/compile ~/workspace/cf-release ~/workspace/diego-release
    
Install Diego

    vagrant ssh
    cd /vagrant
    sudo scripts/install


##Developing
  Development work should be done on the develop branch.
  As a general rule, only CI should commit to master.
