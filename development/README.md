# Lattice Development Readme

## Cluster Development

### Install Dependencies

These instructions assume a basic familiarity with Golang development.
If you have not developed a golang project before, please see the [Golang docs](https://golang.org/doc/)

1. [Golang](https://golang.org/)
1. [Terraform](http://terraform.io)
1. [Vagrant](http://vagrantup.com) with the default Virtualbox provider
1. [Docker](https://docs.docker.com/installation/) (Either a Linux-local dockerd, or on OSX, via [boot2docker](http://boot2docker.io))

### Build Lattice from source

```bash
    $ git clone git@github.com:cloudfoundry-incubator/lattice.git -b develop # may be unstable!
    $ cd lattice
    $ development/setup
    $ development/build
```

### Prepare $PATH and $GOPATH

```bash
    $ source development/env
```

> Note: This will overwrite your $GOPATH.

### Build a local lattice cluster and run smoke tests

```bash
    $ development/run
    $ ltc target 192.168.11.11.xip.io
    $ ltc test -v
```

### Destroy everything

```bash
    $ development/teardown
```

Running `development/setup` subsequent times will clean and re-sync your build directory.

Running `development/teardown` will delete the `build` directory and destroy the running vagrant box. It is destructive and unnecessary
to run `development/teardown` after build.

## Dependency Management

Our continuous deployment pipeline will automatically bump `ltc` dependencies that come from `diego-release`.
There is no need to bump them manually (and you probably shouldn't).
Non-diego-release dependencies should be manually bumped via godep update.
i.e., codegangsta/cli, docker/docker, etc.

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
