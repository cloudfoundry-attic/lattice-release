[![Build Status](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli)

diego-edge-cli
==============

Diego Edge CLI

To use the CLI set the DIEGO_RECEPTOR_ADDRESS environment variable with the location of the diego receptor.

##Commands:

###Start a docker app on diego:


    DIEGO_RECEPTOR_ADDRESS=DIEGO_RECEPTOR_ADDRESS

    diego-edge-cli start APP_NAME --docker-image DOCKER_IMAGE --start-command START_COMMAND

Example Usage:

    DIEGO_RECEPTOR_ADDRESS="http://receptor.192.168.11.11.xip.io" diego-edge-cli start Bingo-app -i "docker:///dajulia3/diego-edge-docker-app" -c "/dockerapp"
