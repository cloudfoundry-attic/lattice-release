#!/bin/bash

set -x -e

export AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY

gem install bundler

OUTPUT=$PWD

cd lattice/concourse/tasks/generate-lattice-bundle-listing

bundler
bundle exec ./generate-listing.rb 
