#!/bin/bash

set -x -e

LATTICE_DIR=$PWD/lattice

docker login --email="$DOCKER_HUB_EMAIL"  --username="$DOCKER_HUB_USERNAME" --password="$DOCKER_HUB_PASSWORD"
docker build --force-rm --no-cache -t $DOCKER_IMAGE $LATTICE_DIR/images/$DOCKER_IMAGE/
docker push $DOCKER_IMAGE
