[![Build Status](https://travis-ci.org/pivotal-cf-experimental/lattice-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/lattice-cli)

lattice-cli
==============

Lattice CLI

##Setup:

Download the appropriate binary for your architecture:

    (MacOS) https://s3-us-west-2.amazonaws.com/lattice/latest/darwin-amd64/ltc
    (Linux) https://s3-us-west-2.amazonaws.com/lattice/latest/linux-amd64/ltc

Make the file executable:

    chmod a+x ltc

And copy the file into your path (e.g., /usr/local/bin) or run it directly with ```./ltc```

##Commands:

###Target a Lattice domain:

    ltc target LATTICE_DOMAIN

###Start a docker app on Lattice:

    ltc start APP_NAME -i DOCKER_IMAGE -- START_COMMAND [APP_ARG1 APP_ARG2...]

###Tail an app's logs on Lattice:

    ltc logs APP_NAME

###Example Usage with Lattice on Vagrant [Lattice](https://github.com/pivotal-cf-experimental/lattice):

    ltc target 192.168.11.11.xip.io
    ltc start lattice-app -i "docker:///cloudfoundry/lattice-app" -- /lattice-app --message="hello"
    ltc logs lattice-app

To view the app in a browser visit http://lattice-app.192.168.11.11.xip.io/

To scale up the app:

    ltc scale lattice-app -i 5

Refresh the browser to see the requests routing to different Docker containers running lattice-app.
