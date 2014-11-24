[![Build Status](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli)

diego-edge-cli
==============

Diego Edge CLI

##Setup:

Make sure you have go installed and your $GOPATH is properly set. Then run:

    go get github.com/pivotal-cf-experimental/diego-edge-cli

This installs the diego-edge-cli binary to $GOPATH/bin.


##Commands:

###Target a diego api:

    diego-edge-cli target API_LOCATION

###Target a diego Loggregator:

    diego-edge-cli target-loggregator DOPPLER_LOCATION

###Start a docker app on diego:

    diego-edge-cli start APP_NAME --docker-image DOCKER_IMAGE --start-command START_COMMAND

###Tail an app's logs on diego:

    diego-edge-cli logs APP_NAME

Example Usage with Diego Edge on Vagrant [Diego Edge](https://github.com/pivotal-cf-experimental/diego-edge):

        diego-edge-cli target http://receptor.192.168.11.11.xip.io
        diego-edge-cli target-loggregator doppler.192.168.11.11.xip.io

        diego-edge-cli start Bingo-app -i "docker:///dajulia3/diego-edge-docker-app" -c "/dockerapp"
        diego-edge-cli logs Bingo-app
