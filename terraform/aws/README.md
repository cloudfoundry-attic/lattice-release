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

Here are some step-by-step instructions for configuring a Lattice cluster via Terraform:

1. Visit the [Lattice GitHub Releases page](https://github.com/cloudfoundry-incubator/lattice/releases#)
2. Select the Lattice version you wish to deploy and download the Terraform example file for your target platform.  The filename will be `lattice.aws.tf`
3. Create an empty folder and place the `lattice.aws.tf` file in that folder.
4. Update the `lattice.aws.tf` by filling in the values for the variables.  Details for the values of those variables are below.

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
* `aws_instance_type_brain`: The machine type to use for the Lattice Brain instance (default `m3.medium`)
* `aws_instance_type_cell`: The machine type to use for the Lattice Cells instances (default `m3.medium`)
* `num_cells`: The number of Lattice Cells to launch (default `3`)
* `lattice_username`: Lattice username (default `user`)
* `lattice_password`: Lattice password (default `pass`)

Refer to the [Terraform AWS provider](https://www.terraform.io/docs/providers/aws/index.html)
documentation for more details about how to configure the proper credentials.

### Deploy

Here are some step-by-step instructions for deploying a Lattice cluster via Terraform:

1. Run the following commands in the folder containing the `lattice.aws.tf` file

  ```bash
  terraform get -update
  terraform apply
  ```

  This will deploy the cluster.

Upon success, terraform will print the Lattice target:

```
Outputs:

  lattice_target = x.x.x.x.xip.io
  lattice_username = xxxxxxxx
  lattice_password = xxxxxxxx
```

which you can use with the Lattice CLI to `ltc target x.x.x.x.xip.io`.

Terraform will generate a `terraform.tfstate` file.  This file describes the cluster that was built - keep it around in order to modify/tear down the cluster.

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
