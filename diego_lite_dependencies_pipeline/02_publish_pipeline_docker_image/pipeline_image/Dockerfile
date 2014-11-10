FROM ubuntu:trusty

ENV HOME /root
RUN \
  echo '' > $HOME/.bashrc; \
  echo "[ -f ~/.bashrc ] && source ~/.bashrc" >> $HOME/.bash_profile

RUN \
  apt-get update; \
  apt-get -y install \
    curl \
    git \
    libcurl4-openssl-dev \
    python-dev \
    python-pip \
    wget \
    jq

RUN \
  bash -l -c "pip install awscli";

RUN \
    git config --global user.email "cf-diego-edge-eng+gocd@pivotal.io"

RUN \
    git config --global user.name "Diego Edge Continuous Deployment"
