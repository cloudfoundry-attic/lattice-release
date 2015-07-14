# Private Docker Registries

Lattice does not currently ship with a private Docker registry.  We plan on remedying this soon to improve our developer experience.  Until then, follow these instructions to spin up a Docker registry on boot2docker, configure the Docker daemon on boot2docker, configure Lattice to allow communication with the private registry, and then import a Docker image to the private registry and launch it via Lattice.

## Launch a Private Docker Registry and configure the Docker daemon via boot2docker

The following assumes:

* Using the latest [http://boot2docker.io/](http://boot2docker.io/)
  * **Note:** There is a bug ([824](https://github.com/boot2docker/boot2docker/issues/824)) with boot2docker v1.7.0, use [v1.6.2](https://github.com/boot2docker/boot2docker/releases/tag/v1.6.2) until v1.7.1 is released.
* Using IP 192.168.59.103 for boot2docker. Find the ip on your boot2docker vm with `boot2docker ip`

**1. Allow the Docker daemon to communicate with the private registry**

Use `boot2docker ssh` to update/create the file `/var/lib/boot2docker/profile` and restart docker:

    sudo su
    echo 'EXTRA_ARGS="--insecure-registry 192.168.59.103:5000"' >> /var/lib/boot2docker/profile
    /etc/init.d/docker restart

**2. Launch the private registry on boot2docker**

From your host machine:

    docker run -p 5000:5000 registry

## Lattice VM configuration

The following assumes you are using the local Vagrant VM.

**Allow Garden-Linux to communicate with the private registry**

SSH to the Lattice VM with `vagrant ssh` from the directory with the Lattice Vagrantfile then modify the Garden-Linux config file (`/etc/init/garden-linux.conf`) and restart Garden-Linux:

    sudo sed -i '15i-insecureDockerRegistryList="192.168.59.103:5000" \\' /etc/init/garden-linux.conf
    sudo initctl stop garden-linux
    sudo initctl start garden-linux

## Push an example docker image to the private registry

Pull the lattice-app image from DockerHub and push it to the local private registry:

    docker pull cloudfoundry/lattice-app
    IMAGE_ID=`docker images | grep lattice-app | awk '{ print $3 }'`
    docker tag $IMAGE_ID 192.168.59.103:5000/lattice-app
    docker push 192.168.59.103:5000/lattice-app

## Launch the docker image hosted on the private registry on Lattice

    ltc create private-lattice-app 192.168.59.103:5000/lattice-app

If there is a problem, run `ltc debug-logs` in a shell while you `ltc remove private-lattice-app` and retry the `ltc create`.

> Note: you will need `ltc` version 0.2.3 or greater.
