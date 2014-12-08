[![Build Status](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli)

diego-edge-cli
==============

Diego Edge CLI

##Setup:

Make sure you have go installed and your $GOPATH is properly set. Then run:

    go get -u github.com/pivotal-cf-experimental/diego-edge-cli

This installs or updates the diego-edge-cli binary to $GOPATH/bin.

##Commands:

###Target a diego domain:

    diego-edge-cli target DIEGO_DOMAIN

###Start a docker app on diego:

    diego-edge-cli start APP_NAME -i DOCKER_IMAGE -- START_COMMAND [APP_ARG1 APP_ARG2...]

###Tail an app's logs on diego:

    diego-edge-cli logs APP_NAME

###Example Usage with Diego Edge on Vagrant [Diego Edge](https://github.com/pivotal-cf-experimental/diego-edge):

    diego-edge-cli target 192.168.11.11.xip.io
    diego-edge-cli start Bingo-app -i "docker:///mylovelyapp" -- /dockerapp --message="hello"
    diego-edge-cli logs Bingo-app
