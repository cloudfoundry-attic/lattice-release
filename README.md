[![Build Status](https://travis-ci.org/pivotal-cf-experimental/lattice-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/lattice-cli)

# Lattice CLI (ltc)

`ltc` provides an easy-to-use command line interface for [Lattice](https://github.com/pivotal-cf-experimental/lattice)

With `ltc` you can:

- `target` a Lattice deployment
- `start`, `scale`, `stop` and `remove` Dockerimage-based applications
- tail `logs` for your running applications
- `list` all running applications and `visualize` their distributions across the Lattice cluster
- fetch detail `status` information for a running application

##Setup:

Download the appropriate binary for your architecture:

Platform | Architecture | Link
-------------------------------
MacOS | amd64 | [https://lattice.s3.amazonaws.com/latest/darwin-amd64/ltc](https://lattice.s3.amazonaws.com/latest/darwin-amd64/ltc)
Linux | amd64 | [https://lattice.s3.amazonaws.com/latest/linux-amd64/ltc](https://lattice.s3.amazonaws.com/latest/linux-amd64/ltc)

Here's a simple installation script.  It assumes `$HOME/bin` is on your $PATH

**Mac**:
```bash
  mkdir -p $HOME/bin
  pushd $HOME/bin
  wget https://lattice.s3.amazonaws.com/latest/darwin-amd64/ltc
  chmod +x ./ltc
  popd
```

**Linux**:
```bash
  mkdir -p $HOME/bin
  pushd $HOME/bin
  wget https://lattice.s3.amazonaws.com/latest/linux-amd64/ltc
  chmod +x ./ltc
  popd
```

#### Installing From Source

You must have [Go](https://golang.org) 1.4+ installed and set up correctly.

```
go get github.com/pivotal-cf-experimental/lattice-cli/ltc
```

## Usage:

`ltc` includes a number of subcommands.  To learn about them:

```
ltc help
ltc help SUBCOMMAND
```

Here are a few key subcommands.

### Target a Lattice cluster:

```
ltc target LATTICE_TARGET
```

When running Lattice locally with Vagrant the default `LATTICE_TARGET` is `192.168.11.11.xip.io`
When deployed to a cloud provider using Terraform you can inspect the resulting `tfstate` file to fetch the `LATTICE_TARGET`

### Start a docker-based app:

```
ltc start APP_NAME DOCKER_IMAGE
```

will start a Dockerimage-based application on Lattice.

We have a simple demo-application that you can play with:

```
ltc start lattice-app cloudfoundry/lattice-app
```

`ltc help start` documents a number of useful options for starting your application.

### Tail an app's logs:

```
ltc logs APP_NAME
```

will start streaming logs emanating from all instances of `APP_NAME`

### See what's running:

```
ltc list
```

Will print out a list of all running applications.

```
ltc status APP_NAME
```

Will print out detailed information about an application.

```
ltc visualize
```

Will print an ascii-art representation of the distribution of containers across the Lattice cluster.

### Example Usage:

    ltc target 192.168.11.11.xip.io
    ltc start lattice-app cloudfoundry/lattice-app
    ltc logs lattice-app

To view the app in a browser visit http://lattice-app.192.168.11.11.xip.io/

To scale up the app:

    ltc scale lattice-app 5

Refresh the browser to see the requests routing to different Docker containers running lattice-app.
