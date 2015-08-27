FROM ubuntu:14.04.3

# install package dependencies
RUN apt-get update -qqy \
  && apt-get -qqy install \
    curl \
    unzip \
    dnsutils \
    make \
    python \
    git \
    mercurial \
    mysql-client \
    fontconfig \
    libjpeg-dev \
    zlib1g-dev \
    libmysqlclient-dev \
    libxml2 \
    libxslt-dev \
    libssl-dev \
  && apt-get clean

# install ruby
ENV RUBY_MAJOR 2.2
ENV RUBY_VERSION 2.2.2
ENV RUBY_DOWNLOAD_SHA256 5ffc0f317e429e6b29d4a98ac521c3ce65481bfd22a8cf845fa02a7b113d9b44
ENV GEM_HOME /usr/local/bundle
ENV PATH $GEM_HOME/bin:$PATH
ENV BUNDLER_VERSION 1.10.5
ENV BUNDLE_APP_CONFIG $GEM_HOME
RUN apt-get -qqy install bison libgdbm-dev ruby autoconf \
  && mkdir -p /usr/src/ruby \
  && curl -fSL -o ruby.tar.gz "http://cache.ruby-lang.org/pub/ruby/$RUBY_MAJOR/ruby-$RUBY_VERSION.tar.gz" \
  && echo "$RUBY_DOWNLOAD_SHA256 *ruby.tar.gz" | sha256sum -c - \
  && tar -xzf ruby.tar.gz -C /usr/src/ruby --strip-components=1 \
  && rm ruby.tar.gz \
  && cd /usr/src/ruby \
  && autoconf \
  && ./configure --disable-install-doc \
  && make -j"$(nproc)" \
  && make install \
  && echo 'gem: --no-rdoc --no-ri' >> "$HOME/.gemrc" \
  && gem install bundler --version "$BUNDLER_VERSION" \
  && bundle config --global path "$GEM_HOME" \
  && bundle config --global bin "$GEM_HOME/bin" \
  && apt-get purge -y --auto-remove bison libgdbm-dev ruby autoconf \
  && apt-get clean \
  && rm -r /usr/src/ruby

# install vagrant
RUN apt-get -qqy install build-essential \
  && curl -L -O "https://dl.bintray.com/mitchellh/vagrant/vagrant_1.6.5_x86_64.deb" \
  && dpkg -i vagrant_1.6.5_x86_64.deb \
  && rm vagrant_1.6.5_x86_64.deb \
  && mkdir -p /home/vcap \
  && HOME=/home/vcap \
  && vagrant plugin install vagrant-aws \
  && vagrant plugin install vagrant-awsinfo \
  && chmod -R 0777 /home/vcap \
  && chmod 0600 /home/vcap/.vagrant.d/insecure_private_key \
  && apt-get purge -y --auto-remove build-essential \
  && apt-get clean

# install bosh cli
RUN gem install bosh_cli --no-rdoc --no-ri

# install bosh-init
RUN curl -L "https://s3.amazonaws.com/bosh-init-artifacts/bosh-init-0.0.43-linux-amd64" > /usr/local/bin/bosh-init \
   && chmod +x /usr/local/bin/bosh-init

# install aws-cli
RUN curl -L -O "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" \
  && unzip awscli-bundle.zip \
  && ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws \
  && rm awscli-bundle.zip

# install golang
RUN cd /usr/local/ \
  && curl -L "https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz" | tar -xz
ENV PATH $PATH:/usr/local/go/bin

# install cf cli
RUN cd /usr/local/bin \
  && curl -L "https://cli.run.pivotal.io/stable?release=linux64-binary&version=6.10.0&source=github-rel" | tar -xz \
  && mkdir -p /usr/local/cf-plugins \
  && curl -L 'https://github.com/cloudfoundry-incubator/diego-cli-plugin/raw/master/bin/linux64/diego-beta.linux64' > /usr/local/cf-plugins/diego-beta \
  && chmod +x /usr/local/cf-plugins/diego-beta

# install interpolate command
RUN curl "https://s3.amazonaws.com/cfla-dropbox/interpolate" > /usr/local/bin/interpolate \
  && curl -L "https://s3.amazonaws.com/cfla-dropbox/interpolate-yaml" > /usr/local/bin/interpolate-yaml \
  && chmod +x /usr/local/bin/interpolate-yaml \
  && chmod +x /usr/local/bin/interpolate

# install fly
RUN curl -L "https://cfla.ci.cf-app.com/api/v1/cli?arch=amd64&platform=linux" > /usr/local/bin/fly \
  && chmod +x /usr/local/bin/fly

# install spiff
RUN curl -L -O "https://github.com/cloudfoundry-incubator/spiff/releases/download/v1.0.7/spiff_linux_amd64.zip" \
  && unzip spiff_linux_amd64.zip \
  && rm spiff_linux_amd64.zip \
  && mv spiff /usr/local/bin/spiff \
  && chmod +x /usr/local/bin/spiff

# install jq
RUN curl -L "http://stedolan.github.io/jq/download/linux64/jq" > /usr/local/bin/jq \
  && chmod +x /usr/local/bin/jq
