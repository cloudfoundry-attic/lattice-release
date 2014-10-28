##Running the box
   
    vagrant up
    
The box lives at 192.168.11.11
    
etcd is accessible on port 4001
Loggregator is accessible via loggregator.192.168.11.11.xip.io
    
##To provision a box from scratch:
    
Compile the necessary binaries (Currently this is run outside of the guest vm)

    ./compile
    
Install Diego

    vagrant ssh
    cd /vagrant
    sudo ./install
