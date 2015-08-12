#Lattice Development Readme

##Cluster Development

###Install Dependencies
These instructions assume a basic familiarity with Golang development.
If you have not developed a golang project before, please see the [Golang docs](https://golang.org/doc/)

####Install:

1. [Golang](https://golang.org/)
1. [Terraform](http://terraform.io)
1. [Vagrant](http://vagrantup.com) with the default Virtualbox provider
1. [Docker](https://docs.docker.com/installation/)
1. Setup source code dependencies, the ltc cli, and environment

```bash
    $ mkdir -p ~/workspace
    $ cd ~/workspace
    $ git clone git@github.com:cloudfoundry/cf-release.git --recursive
    $ git clone git@github.com:cloudfoundry-incubator/diego-release.git
    $ cd diego-release/scripts && ./update && cd -
    $ git clone git@github.com:cloudfoundry-incubator/cf-routing-release.git
    $ cd cf-routing-release/scripts && ./update && cd -
    $ export GOPATH=~/workspace/diego-release
    $ export PATH="$PATH:~/workspace/diego-release/bin"
    $ go get github.com/dajulia3/godep #Our forked version of godep that handles submodules:
    $ go get github.com/onsi/ginkgo/ginkgo
    $ go get github.com/onsi/gomega
    $ go get github.com/maxbrunsfeld/counterfeiter
    $ cd ~/workspace/diego-release/src/github.com/cloudfoundry-incubator
    $ git clone git@github.com:cloudfoundry-incubator/lattice.git
    $ cd ~/workspace/diego-release/src/github.com/cloudfoundry-incubator/lattice
    $ git checkout develop #Develop is not guaranteed to be stable, but you're a contributor, so you're awesome enough to handle it!
    $ DIEGO_VERSION=$(cat DIEGO_VERSION) cd $GOPATH && git checkout $DIEGO_VERSION && git submodule update --init --recursive && cd -
    $ rm -rf $GOPATH/src/github.com/docker/docker && go get github.com/docker/docker
    $ go get ./...
    $ cd ltc
    $ godep restore
    $ go install
    $ export DOCKER_IMAGE=cloudfoundry/lattice-pipeline
    $ $(boot2docker shellinit)
    $ docker pull cloudfoundry/lattice-pipeline
    $ docker pull cloudfoundry/lattice-app
```

#### Add our helpful aliases to your bash_profile

```bash
    $ export PULL_DOCKER_IMAGE=false
    $ alias recompile-lattice="cd ~/workspace && $GOPATH/src/github.com/cloudfoundry-incubator/lattice/pipeline/helpers/run_with_docker /workspace/diego-release/src/github.com/cloudfoundry-incubator/lattice/pipeline/01_compilation/compile_lattice_tar && mv -v ./lattice.tgz $GOPATH/src/github.com/cloudfoundry-incubator/lattice/"
    $ alias remake-vagrant="cd-lattice; vagrant destroy --force; recompile-lattice && VAGRANT_LATTICE_TAR_PATH=/vagrant/lattice.tgz vagrant up --provider=virtualbox; go install github.com/cloudfoundry-incubator/lattice/ltc"
```

#### Build a vagrant vm-deployed lattice cluster, verify that it's up and running:

```bash
    $ remake-vagrant
    $ ltc test -v
```
## ltc Development

If you only want to develop against ltc, but do not care to make local cluster changes,
a viable and more stable option is to clone lattice and godep restore from the vendored Godeps.

```bash
    $ mkdir -p ~/workspace/go
    $ export GOPATH=~/workspace/go
    $ export PATH="$PATH:$GOPATH/bin"
    $ go get github.com/dajulia3/godep #Our forked version of godep that handles submodules:
    $ go get github.com/onsi/ginkgo/ginkgo
    $ go get github.com/onsi/gomega
    $ go get github.com/maxbrunsfeld/counterfeiter
    $ mkdir -p $GOPATH/src/github.com/cloudfoundry-incubator
    $ cd $GOPATH/src/github.com/cloudfoundry-incubator
    $ git clone git@github.com:cloudfoundry-incubator/lattice.git
    $ cd lattice/ltc
    $ go get ./...
    $ godep restore
    $ go install
    $ ./scripts/test #run the unit tests!
```

## Dependency Management

Our continuous deployment pipeline should automatically bump LTC dependencies that come from diego-release.
There is no need to bump them manually (and you probably shouldn't).
Non-diego-release dependencies should be manually bumped via godep update.
i.e., codegangsta/cli, cloudfoundry/noaa, docker/docker, etc.

For any given SHA of lattice, ltc should use the same diego-release dependencies that that lattice cluster tar builds against.
This keeps any overlapping dependencies in lock step.
That is, if the version of receptor in the lattice cluster gets updated, the receptor client that ltc uses will be updated with it.
The pipeline ensures this on master. Nesting lattice in the GOPATH constituted by diego-release ensures this locally.

## Pull Requests

Please make pull requests against the develop branch.
We will not accept untested changes. Please test-drive your code.

## Branches/Stability

We strive to keep master stable.
Tagged releases are vetted releases based on master.
We make no guarantees about the stability of the develop branch on a per-commit basis.
