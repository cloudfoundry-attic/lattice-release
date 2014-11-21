[![Build Status](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli)

diego-edge-cli
==============

Diego Edge CLI

##Commands:

###Target a diego api:

    diego-edge-cli target API_LOCATION

###Target a diego Loggregator:

    diego-edge-cli target-loggregator DOPPLER_LOCATION

###Start a docker app on diego:

    diego-edge-cli start APP_NAME --docker-image DOCKER_IMAGE --start-command START_COMMAND

###Tail an app's logs on diego:

    diego-edge-cli logs APP_NAME

Example Usage:

        diego-edge-cli target receptor.192.168.11.11.xip.io
        diego-edge-cli target-loggregator doppler.192.168.11.11.xip.io

        diego-edge-cli start Bingo-app -i "docker:///dajulia3/diego-edge-docker-app" -c "/dockerapp"
        diego-edge-cli logs Bingo-app
