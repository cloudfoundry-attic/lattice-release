# Getting Started

You can run Lattice easily on your laptop with a [Vagrant VM](https://github.com/cloudfoundry-incubator/lattice#local-deployment) or deploy a [cluster of machines](https://github.com/cloudfoundry-incubator/lattice#clustered-deployment) with [AWS](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/aws/README.md), [Digital Ocean](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/digitalocean/README.md), [Google Cloud](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/google/README.md) or [Openstack](https://github.com/cloudfoundry-incubator/lattice/blob/master/terraform/openstack/README.md).

This tutorial walks you through the Vagrant VM flow and using Lattice to start applications based on a docker image, scale them up and down and retrieve logs.

## Pre-Requisites for the Vagrant VM

- [Vagrant](https://www.vagrantup.com)
- [VirtualBox](https://www.virtualbox.org) or [VMware Fusion](http://www.vmware.com/products/fusion) (version 7 or later)
- [git](https://git-scm.com)

## Starting the Lattice Vagrant VM

Consult the [GitHub Releases](https://github.com/cloudfoundry-incubator/lattice/releases) page, and decide on which version of Lattice you plan to use. We also offer [nightly builds](https://lattice.s3.amazonaws.com/nightly/index.html) for cutting-edge functionality, but recommend the releases for stability. 

First, download one of the bundles listed on the above pages.  Then, unzip it and change into the `vagrant` folder of the unpacked bundle.

    unzip lattice-bundle-VERSION-PLATFORM.zip
    cd lattice-bundle-VERSION-PLATFORM/vagrant
    vagrant up

Then bring up the Vagrant box:

**Virtualbox**:

    vagrant up --provider virtualbox

**VMware Fusion**:

    vagrant up --provider vmware_fusion

The VM should download and start.

> By default the Lattice VM will be reachable at `192.168.11.11`. You can set the `LATTICE_SYSTEM_IP` environment variable when running `vagrant up` to modify this.  

> If you are trying to run both the VirtualBox and VMWare providers on the same machine, you'll need to run them on different private networks (subnets) that do not conflict.

> Learn more about deploying Lattice at the GitHub [README](https://github.com/cloudfoundry-incubator/lattice/tree/v0.4.0)

## `ltc` - the Lattice CLI

The downloaded bundle also includes the command-line tool `ltc`.  This can be copied to a folder in your `PATH` or executed in-place (e.g., `./ltc`).

Further instructions can be found [here](https://github.com/cloudfoundry-incubator/lattice/tree/v0.4.0/ltc).

## Targeting Lattice

You need to tell `ltc` how to connect to your Lattice deployment.  The target domain should be printed out when you `vagrant up`.  If you have not changed the default settings you can:

    ltc target 192.168.11.11.xip.io

## Launching and Routing to a Container

We have a simple Go-based demo web application hosted on the Docker registry at [`cloudfoundry/lattice-app`](https://registry.hub.docker.com/u/cloudfoundry/lattice-app).  You can launch this image by running:

    ltc create lattice-app cloudfoundry/lattice-app

Once the application is running, `ltc` will emit the route you can use to access the application:

    Starting App: lattice-app...
    lattice-app is now running.
    http://lattice-app.192.168.11.11.xip.io

You should be able to visit `lattice-app.192.168.11.11.xip.io` in your browser.

The `lattice-app` has three endpoints:

- `/` is a pretty landing page that includes the instance's index and uptime
- `/env` prints out the instance's environment
- `/exit` causes the instance to crash

## Tailing Logs

To stream logs from your running `lattice-app`:

    ltc logs lattice-app

Visiting `lattice-app.192.168.11.11.xip.io` will emit log messages that should be visible in your terminal.

## Listing Applications

To view a list of all running applications:

    ltc list

## Scaling Applications

To scale `lattice-app` to 3 instances:

    ltc scale lattice-app 3

Now `ltc list` should show that `3/3` instances are running and `ltc logs lattice-app` will aggregate logs from all three instances.

Visiting `lattice-app.192.168.11.11.xip.io` should cycle through the different instances that are running.  Each instance will have a unique index.

## Getting Application Details

To view detailed information about your running instances:

    ltc status lattice-app

## Visualizing Containers

To visualize the distribution of containers on your lattice cluster:

    ltc visualize

If you deploy a cluster of Lattice cells `ltc visualize` will show you the distribution of instances across the cluster.

## Crash Recovery Demo

Visit `lattice-app.192.168.11.11.xip.io/exit`

This will cause one of the `lattice-app` instances to exit.  Lattice will immediately restart the instance.

If you cause an instance of `lattice-app` to exit repeatedly Lattice will eventually start applying a backoff policy and restart the instance only after increasing intervals of time (30s, 60s, etc...)

## Where to go from here:

- push your own Docker image
- learn more about [`ltc`](/docs/ltc.md)
- learn more about the RESTful [`Lattice API`](/docs/lattice-api.md).  This allows you to launch one off tasks in addition to long running processes.
