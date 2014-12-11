#Vagrant

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
     Lattice is now installed and running. You may target it with the Lattice cli via:
     192.168.194.130.xip.io
     
     $ ltc target 192.168.194.130.xip.io 
     

##Updating

Currently, Diego Edge does not support updating via provision.
So to update, you have to destroy the box and bring it back up as shown below:

  vagrant destroy --force
  git pull
  vagrant up

####Running Vagrant with a custom lattice tar

    VAGRANT_LATTICE_TAR_PATH=/vagrant/lattice.tgz vagrant up


#Aws

Follow [Amazon's instructions](http://docs.aws.amazon.com/cli/latest/userguide/installing.html) for setting up the aws cli.

Configure the aws cli with your aws access key, aws secret access key, and the us-west-1 region
   
     aws config
   
Set up security group

     aws ec2 create-security-group --group-name lattice --description "lattice security group." 
   
Open up the instance to incoming tcp traffic
    
     aws ec2 authorize-security-group-ingress --group-name lattice --protocol tcp --port 1-65535 --cidr 0.0.0.0/0
     
Creates a credentials file containing the username and password that you want to use for the cli
     
     echo "RECEPTOR_USERNAME=<Your Username>" > lattice-credentials
     echo "RECEPTOR_PASSWORD=<Your Password>" >> lattice-credentials

Launch an instance of lattice with your base64 encoded username and password file

    aws ec2 run-instances --image-id ami-67485a22 --security-groups default --key-name ec2-west-1 --user-data `base64 lattice-credentials`
    
#Testing the Diego Edge Box

 Follow the [whetstone instructions](https://github.com/pivotal-cf-experimental/whetstone) for lattice

#Using Diego Edge

 Use the [Lattice Cli](https://github.com/pivotal-cf-experimental/lattice-cli).


#Developing
  Development work should be done on the develop branch.
  As a general rule, only CI should commit to master.
