# Lattice Development Read-Me

## Continuous Integration

Our [Concourse](http://concourse.ci) CI system is available at [ci.lattice.cf](https://ci.lattice.cf).

## Cluster Development

### Install Dependencies

These instructions assume a basic familiarity with Golang development.
If you have not developed a golang project before, please see the [Golang docs](https://golang.org/doc/)

1. [Golang](https://golang.org/)
1. [Terraform](http://terraform.io)
1. [Vagrant](http://vagrantup.com) with the default Virtualbox provider
1. [Docker](https://docs.docker.com/installation/) (Either a Linux-local dockerd, or on OSX, via [Docker Machine](https://docs.docker.com/installation/mac/))

### Build Lattice from source

```bash
    $ git clone git@github.com:cloudfoundry-incubator/lattice.git
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

`ltc` is built using the vendored dependencies under `Godeps`.  These need to match the versions used in your Lattice cluster, which are tracked as submodules of the diego-release project.  Non-Diego dependencies (i.e., `codegangsta/cli`, `docker/docker`) should be manually bumped via `godep update`.  

## Pull Requests

Please make pull requests against the master branch.
We will not accept untested changes. Please test-drive your code.

## Branches/Stability

[GitHub Releases](https://github.com/cloudfoundry-incubator/lattice/releases) have passed our CI pipeline and should be stable.  Bug reports / pull requests are always welcome.

We make no guarantees about the stability of the master branch on a per-commit basis.  


