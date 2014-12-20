[![Build Status](https://travis-ci.org/pivotal-cf-experimental/lattice-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/lattice-cli)

lattice-cli
==============

Lattice CLI

##Setup:

Make sure you have go installed and your $GOPATH is properly set. Then run:

    go get -u github.com/pivotal-cf-experimental/lattice-cli/ltc

This installs or updates the `ltc` binary to $GOPATH/bin.

##Commands:

###Target a Lattice domain:

    ltc target LATTICE_DOMAIN

###Start a docker app on Lattice:

    ltc start APP_NAME -i DOCKER_IMAGE -- START_COMMAND [APP_ARG1 APP_ARG2...]

###Tail an app's logs on Lattice:

    ltc logs APP_NAME

###Example Usage with Lattice Edge on Vagrant [Diego Edge](https://github.com/pivotal-cf-experimental/diego-edge):

    ltc target 192.168.11.11.xip.io
    ltc start Bingo-app -i "docker:///cloudfoundry/lattice-app" -- /lattice-app --message="hello"
    ltc logs Bingo-app
