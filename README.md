[![Build Status](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli)

diego-edge-cli
==============

Diego Edge CLI

##Commands:

###Target a diego api:

    diego-edge-cli target API_LOCATION


###Start a docker app on diego:

    diego-edge-cli start APP_NAME --docker-image DOCKER_IMAGE --start-command START_COMMAND

Example Usage:

        diego-edge-cli target receptor.192.168.11.11.xip.io
        diego-edge-cli start Bingo-app -i "docker:///dajulia3/diego-edge-docker-app" -c "/dockerapp"
