#Vagrant

##Running the box

    $ vagrant up

Vagrant up spins up a virtual hardware environment that is accessible at 192.168.11.11. You can do this with either VMware Fusion or VirtualBox:. You can specify your preferred provider with the --provider flag.

Virtualbox:

     $ vagrant up --provider virtualbox

VMware Fusion:

     $ vagrant up --provider vmware_fusion

## Networking Conflicts
If you are trying to run both the Virtual Box and VMWare providers on the same machine, 
you'll need to run them on different private networks that do not conflict. 

Change the line of the vagrant file that says

      config.vm.network "private_network", ip: "192.168.11.11"

to 

      config.vm.network "private_network", ip: "192.168.80.100"

or a different non-conflicting IP.

The output from vagrant up will provide instructions on how to target Lattice. 

Use the [Lattice Cli](https://github.com/pivotal-cf-experimental/lattice-cli) to target Lattice.

##Example Usage

     $ vagrant up
     
     Bringing machine 'default' up with 'vmware_fusion' provider...
     ...
     ...
     Lattice is now installed and running. You may target it with the Lattice cli via:
     192.168.194.130.xip.io
     
     $ ltc target 192.168.194.130.xip.io 
     

##Updating

Currently, Lattice does not support updating via provision.
So to update, you have to destroy the box and bring it back up as shown below:

     vagrant destroy --force
     git pull
     vagrant up
  
##Troubleshooting
-  xip.io is sometimes flaky, resulting in no such host errors.
-  The alternative that we have found is to use dnsmasq configured to resolve all xip.io addresses to 192.168.11.11.
-  This also requires creating a /etc/resolvers/io file that points to 127.0.0.1. See further instructions [here] (http://passingcuriosity.com/2013/dnsmasq-dev-osx/). 

##Running Vagrant with a custom lattice tar

    VAGRANT_LATTICE_TAR_PATH=/vagrant/lattice.tgz vagrant up

#Aws

##Setting up AWS With a Collocated Installation
Follow [Amazon's instructions](http://docs.aws.amazon.com/cli/latest/userguide/installing.html) for setting up the aws cli.

Configure the aws cli with your aws access key, aws secret access key, and the us-west-1 region

     aws config

Set up security group

     aws ec2 create-security-group --group-name lattice --description "lattice security group."

Open up the instance to incoming tcp traffic

     aws ec2 authorize-security-group-ingress --group-name lattice --protocol tcp --port 1-65535 --cidr 0.0.0.0/0

Creates a credentials file containing the username and password that you want to use for the cli

     echo "LATTICE_USERNAME=<Your Username>" > lattice-credentials
     echo "LATTICE_PASSWORD=<Your Password>" >> lattice-credentials

Create a key pair

    aws ec2 create-key-pair --key-name lattice-key

Launch an instance of lattice with your base64 encoded username and password file

    aws ec2 run-instances --image-id ami-03958746 --security-groups lattice --key-name lattice-key --user-data `base64 lattice-credentials`

Find the PublicIpAddress of the instance you just launched.  You can either use the EC2 Web Console or run the following
command that lists all instances provisioned with the above AMI.

    aws ec2 describe-instances --filter "Name=image-id,Values=ami-03958746" | egrep -i "reservationid|instanceid|imageid|publicipaddress|launchtime"

Sample output:

    aws ec2 describe-instances --filter "Name=image-id,Values=ami-03958746" | egrep -i "reservationid|instanceid|imageid|publicipaddress|launchtime"
            "ReservationId": "r-68fb47a2",
                    "LaunchTime": "2014-12-16T15:43:06.000Z",
                    "PublicIpAddress": "12.345.130.132",
                    "InstanceId": "i-d2b59718",
                    "ImageId": "ami-03958746",

Target Lattice using the [Lattice Cli](https://github.com/pivotal-cf-experimental/lattice-cli). The target will be the PublicIpAddress with the suffix "xip.io" appended. The cli will prompt for the username and password used above.

    ltc target 12.345.130.132.xip.io
        Username: <Your Username>
        Password: <Your Password>


##Setting up AWS With separate Coordinator and Cell(s)

1. Follow [Amazon's instructions](http://docs.aws.amazon.com/cli/latest/userguide/installing.html) for setting up the aws cli. Configure the aws cli with your aws access key, aws secret access key, and the us-west-1 region. 
   
   ```
   aws config
   ```
   
1. Create a Virtual Private Cloud, VPC. Find the VpcId from the output.
   
   ```
   aws ec2 create-vpc --cidr-block 10.10.0.0/16
   ```
     
1. Create a Subnet with the VpcId created above. Find the SubnetId from the output.

   ```
   aws ec2 create-subnet --vpc-id vpc-XXXXXXXX --cidr-block 10.10.1.0/24
   ```

1. Configure the subnet to assign public IP addresses on launch.
    
   ``` 
   aws ec2 modify-subnet-attribute --subnet-id subnet-XXXXXXXX --map-public-ip-on-launch
   ```
    
1. Create an Internet Gateway. Find the InternetGatewayId from the output.
    
   ```
    aws ec2 create-internet-gateway    
   ```
    
1. Attach the internet gateway to the VPC created above. Use the InternetGatewayId and the VpcId from above.
    
   ```
    aws ec2 attach-internet-gateway --internet-gateway-id igw-XXXXXXXX --vpc-id vpc-XXXXXXXX
   ```

1. Create a new routing table associated with the VPC.

   ```
    aws ec2 create-route-table --vpc-id vpc-XXXXXXXX
   ```
1. Associate the route table with the subnet.

   ```
    aws ec2 associate-route-table --subnet-id=subnet-XXXXXXXX --route-table-id=rtb-XXXXXXXX
   ```
    
1. Define a default route via the internet gateway on the routing table.
       
   ```
    aws ec2 create-route --route-table-id rtb-XXXXXXXX --destination-cidr-block 0.0.0.0/0 --gateway-id igw-XXXXXXXX   
   ```
   
1. Create a security group. Use the VpcId from above. Find the GroupId of new security group from the output.

   ```
    aws ec2 create-security-group --group-name lattice --description "lattice security group." --vpc-id vpc-XXXXXXXX    
   ```

1. Open up the instance to incoming traffic. Use the GroupId from above.
    
   ```
    aws ec2 authorize-security-group-ingress --group-id sg-XXXXXXXX --protocol tcp --port 1-65535 --cidr 0.0.0.0/0
    aws ec2 authorize-security-group-ingress --group-id sg-XXXXXXXX --protocol udp --port 1-65535 --cidr 0.0.0.0/0
   ```

1. Creates a credentials file containing the username and password that you want to use for the cli
     
   ```
    echo "LATTICE_USERNAME=<Your Username>" > lattice-credentials
    echo "LATTICE_PASSWORD=<Your Password>" >> lattice-credentials
   ```

1. Create a key pair

   ```
    aws ec2 create-key-pair --key-name lattice-key
   ```
      
1. Launch an instance of the lattice coordinator. This uses the base64 encoded username and password credentials file from step 8. It also launches the instance with a private ip address of 10.10.1.11. Use the SubnetId and GroupId from above.

   ```
    aws ec2 run-instances \
        --subnet-id subnet-XXXXXXXX \
        --security-group-ids sg-XXXXXXXX \
        --image-id ami-6909152c \
        --private-ip-address 10.10.1.11 \
        --key-name lattice-key3 \
        --user-data `base64 lattice-credentials`
   ```

1. Creates a diego-cell-configuration file containing the private ip address from step 10.
        
   ```
    echo "LATTICE_COORDINATOR_IP=10.10.1.11" > diego-cell-config
   ```

1. Launch at least one instance of the diego-cell.

   ```
    aws ec2 run-instances \
     --subnet-id subnet-XXXXXXXX \
     --security-group-ids sg-XXXXXXXX \
     --image-id ami-73091536 \
     --key-name lattice-key3 \
     --user-data `base64 diego-cell-config`
   ```

##Targeting Lattice on AWS

Find the PublicIpAddress of the lattice coordinator instance you just launched.  You can either use the EC2 Web Console or run the following 
command that lists all instances provisioned with the above AMI.
   
   ```
    aws ec2 describe-instances --filter "Name=image-id,Values=ami-8fb8aaca" | egrep -i "reservationid|instanceid|imageid|publicipaddress|launchtime"
   ```
   
Sample output with a PublicIpAddress of 12.345.130.132:
   
   ```
    aws ec2 describe-instances --filter "Name=image-id,Values=ami-8fb8aaca" | egrep -i "reservationid|instanceid|imageid|publicipaddress|launchtime"
            "ReservationId": "r-68fb47a2",
                    "LaunchTime": "2014-12-16T15:43:06.000Z",
                    "PublicIpAddress": "12.345.130.132",
                    "InstanceId": "i-d2b59718",
                    "ImageId": "ami-8fb8aaca",
   ```
      
Target Lattice using the [Lattice Cli](https://github.com/pivotal-cf-experimental/lattice-cli). The target will be the PublicIpAddress with the suffix "xip.io" appended. The cli will prompt for the username and password used above.
 ```
  ltc target 12.345.130.132.xip.io
      Username: <Your Username>
      Password: <Your Password>
 ```
   
#Testing the Lattice Box

 Follow the [whetstone instructions](https://github.com/pivotal-cf-experimental/whetstone) for lattice

#Using Lattice

 Use the [Lattice Cli](https://github.com/pivotal-cf-experimental/lattice-cli).


#Developing
  Development work should be done on the develop branch.
  As a general rule, only CI should commit to master.
  
