# Lattice Terraform templates for Amazon Web Services

This project contains [Terraform](https://www.terraform.io/) templates to help you deploy
[Lattice](https://github.com/cloudfoundry-incubator/lattice) on
[Amazon Web Services](http://aws.amazon.com/).

## Usage

### Prerequisites

* An [Amazon Web Services account](http://aws.amazon.com/)
* An [AWS Access and Secret Access Keys](http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSGettingStartedGuide/AWSCredentials.html)
* An [AWS EC2 Key Pairs](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)

### Configure

Create a `lattice.tf` file (or use the provided [example](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/aws/lattice.tf.example)) and add the following contents updating the variables properly:

```
module "lattice-aws" {
    source = "github.com/cloudfoundry-incubator/lattice/terraform/aws"

    # AWS access key
    aws_access_key = "<CHANGE-ME>"

    # AWS secret key
    aws_secret_key = "<CHANGE-ME>"

    # The SSH key name to use for the instances
    aws_key_name = "<CHANGE-ME>"

    # Path to the SSH private key file
    aws_ssh_private_key_file = "<CHANGE-ME>"

    # The number of Lattice Cells to launch
    num_cells = "3"

    #################################
    ###  Optional Settings Below  ###
    #################################

    #If you wish to use your own lattice release instead of the latest version, uncomment the variable assignment below
    #and set it to your own lattice tar's path.
    # local_lattice_tar_path = ~/lattice.tgz

    # AWS region (e.g., us-west-1)
    # aws_region = "<CHANGE-ME>"
}
```

The available variables that can be configured are:

* `aws_access_key`: AWS access key
* `aws_secret_key`: AWS secret key
* `aws_key_name`: The SSH key name to use for the instances
* `aws_ssh_private_key_file`: Path to the SSH private key file
* `aws_ssh_user`: SSH user (default `ubuntu`)
* `aws_region`: AWS region (default `us-east-1`)
* `aws_vpc_cidr_block`: The IPv4 address range that machines in the network are assigned to, represented as a CIDR block (default `10.0.0.0/16`)
* `aws_subnet_cidr_block`: The IPv4 address range that machines in the network are assigned to, represented as a CIDR block (default `10.0.1.0/24`)
* `aws_image`: The name of the image to base the launched instances (default `ubuntu trusty 64bit hvm ami`)
* `aws_instance_type_coordinator`: The machine type to use for the Lattice Coordinator instance (default `m3.medium`)
* `aws_instance_type_cell`: The machine type to use for the Lattice Cells instances (default `m3.medium`)
* `num_cells`: The number of Lattice Cells to launch (default `3`)
* `lattice_username`: Lattice username (default `user`)
* `lattice_password`: Lattice password (default `pass`)

Refer to the [Terraform AWS provider](https://www.terraform.io/docs/providers/aws/index.html)
documentation for more details about how to configure the proper credentials.

### Deploy

Get the templates and deploy the cluster:

```
terraform get -update
terraform apply
```

After the cluster has been successfully, terraform will print the Lattice domain:

```
Outputs:

  lattice_target = x.x.x.x.xip.io
  lattice_username = xxxxxxxx
  lattice_password = xxxxxxxx
```

### Use

Refer to the [Lattice CLI](https://github.com/cloudfoundry-incubator/lattice/tree/master/ltc) documentation.

### Destroy

Destroy the cluster:

```
terraform destroy
```

## Copyright

See [LICENSE](https://github.com/cloudfoundry-incubator/lattice/blob/master/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
