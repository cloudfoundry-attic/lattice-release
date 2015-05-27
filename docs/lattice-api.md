# The Lattice API

`ltc` is simply a wrapper around a Lattice's powerful RESTful API.  Whereas `ltc` is geared towards launching Docker images on Lattice, you can use the underlying API to construct and request more elaborate container workloads.

The Lattice API is provided by Diego's Receptor component.  You can learn about the API [here](https://github.com/cloudfoundry-incubator/receptor/blob/master/doc/README.md).  We recommend using the Golang client provided by the [Receptor](https://github.com/cloudfoundry-incubator/receptor).

There are a handful of provisos to keep in mind when using the Diego Receptor API with Lattice:

- Lattice does not ship with a default rootfs.  As such you must specify a Docker image for a rootfs.  One option is to use something lightweight like [busybox](https://registry.hub.docker.com/_/busybox/).  To use busybox set `rootfs: "docker:///library/busybox"` when constructing DesiredLRPs and Tasks.  
 
	If you need a more complete Linux distribution you can use the rootfs that ships with Cloud Foundry ([lucid64](https://registry.hub.docker.com/u/cloudfoundry/lucid64/), [trusty64](https://registry.hub.docker.com/u/cloudfoundry/trusty64/)).  Simply set `rootfs: "docker:///cloudfoundry/lucid64"` or `rootfs: "docker:///cloudfoundry/trusty64"`.  Be warned that these are quite large: the first container to launch on a given Cell will spend a while downloading the image.

- Lattice does not apply any firewall rules to containers: any container talk to any other container (or any other endpoint within the VPC for that matter).  As such, there is no need to specify `egress_rules` for your Tasks and DesiredLRPs.

- It is important to understand how [Lattice works with Docker images](/docs/troubleshooting.md#how-does-lattice-work-with-docker-images).  In short, as a consumer of the Lattice API it is your responsiblity to specify which command Lattice should run after mounting the Docker image file system.  `ltc` populates this information in the DesiredLRP by fetching the metadata associated with the Docker image from the Docker registry.