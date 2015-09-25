#!/bin/bash

set -ex

export AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY
gem install bundler

pushd lattice/concourse/tasks/generate-lattice-bundle-listing
  bundle install
  bundle exec ./generate-listing.rb
popd
