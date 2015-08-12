FROM ubuntu:14.04.2

ENV HOME /root
ENV PATH /usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
ENV GOROOT /usr/local/go
ENV GO_VERSION 1.4.2
ENV TERRAFORM_VERSION 0.6.3

RUN \
  apt-get update; \
  apt-get -y upgrade; \
  apt-get -y install \
    software-properties-common \
    curl \
    libcurl4-openssl-dev \
    python-dev \
    python-pip \
    wget \
    zip \
    unzip \
    jq \
    ruby \
    ruby-dev; \
  apt-get -y install \
    asciidoc xmlto --no-install-recommends; \
  apt-get -y install \
    pkg-config \
    autoconf; \
  apt-get -y install \
    uuid-dev \
    libattr1-dev \
    zlib1g-dev \
    libacl1-dev \
    e2fslibs-dev \
    libblkid-dev \
    liblzo2-dev; \
  add-apt-repository ppa:git-core/ppa -y; \
  apt-get update; \
  apt-get -y install git; \
  gem install json

RUN \
  bash -l -c "pip install awscli"; \
  wget --quiet https://dl.bintray.com/mitchellh/terraform/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
  unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/terraform; \
  rm -v terraform_${TERRAFORM_VERSION}_linux_amd64.zip; \
  cd /usr/local/bin && ln -sv /usr/local/terraform/* .

RUN \
  wget https://storage.googleapis.com/golang/go${GO_VERSION}.src.tar.gz  -O - | tar -xzf - -C /usr/local && \
  cd /usr/local/go/src && \
  GOOS=linux GOARCH=amd64 CGO_ENABLED=1 ./make.bash --no-clean && \
  GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 ./make.bash --no-clean; \
  mkdir -p ~/go/src; \
  GOBIN=/usr/local/bin GOPATH=~/go go get -v github.com/dajulia3/godep; \
  rm -rf ~/go

RUN \
  echo "Host github.com" > /etc/ssh/ssh_config && \
  echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config; \
  echo '' > $HOME/.bashrc; \
  echo "[ -f ~/.bashrc ] && source ~/.bashrc" >> $HOME/.bash_profile; \
  git config --global user.email "cf-diego-edge-eng+gocd@pivotal.io" && \
  git config --global user.name "Lattice Continuous Deployment";

