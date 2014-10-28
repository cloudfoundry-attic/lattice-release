##Running the box
   
    vagrant up
    
The box lives at 192.168.11.11
    
etcd is accessible on port 4001
Loggregator is accessible via loggregator.192.168.11.11.xip.io
    
##To provision a box from scratch:
    
Clone Diego-Lite and its dependencies cf-release and diego-lite and compile the necessary binaries (Currently this is run outside of the guest vm)
   
    mkdir -p ~/workspace
    cd workspace
    git clone git@github.com:pivotal-cf-experimental/diego-lite.git
    git clone git@github.com:cloudfoundry-incubator/diego-release.git
    git clone git@github.com:cloudfoundry/cf-release.git
   
    cd diego-lite
    ./compile ~/workspace/cf-release ~/workspace/diego-release
    
Install Diego

    vagrant ssh
    cd /vagrant
    sudo ./install


##Developing
  Development work should be done on the develop branch.
  As a general rule, only CI should commit to master.
