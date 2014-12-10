[![Build Status](https://travis-ci.org/pivotal-cf-experimental/lattice-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/lattice-cli)

lattice-cli
==============

Lattice CLI

##Setup:

Make sure you have go installed and your $GOPATH is properly set. Then run:

    go get -u github.com/pivotal-cf-experimental/lattice-cli

This installs or updates the lattice-cli binary to $GOPATH/bin.

##Commands:

###Target a Lattice domain:

    lattice-cli target LATTICE_DOMAIN

###Start a docker app on Lattice:

    lattice-cli start APP_NAME -i DOCKER_IMAGE -- START_COMMAND [APP_ARG1 APP_ARG2...]

###Tail an app's logs on Lattice:

    lattice-cli logs APP_NAME

###Example Usage with Lattice Edge on Vagrant [Diego Edge](https://github.com/pivotal-cf-experimental/diego-edge):

    lattice-cli target 192.168.11.11.xip.io
    lattice-cli start Bingo-app -i "docker:///diegoedge/diego-edge-docker-app" -- /dockerapp --message="hello"
    lattice-cli logs Bingo-app
