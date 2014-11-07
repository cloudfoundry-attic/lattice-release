FROM ubuntu:trusty

ENV HOME /root
RUN \
  echo "[ -f ~/.bashrc ] && source ~/.bashrc" >> $HOME/.bash_profile

RUN \
  apt-get update; \
  apt-get -y install \
    curl \
    git \
    libcurl4-openssl-dev \
    python-pip \
    wget \
    jq

RUN \
  bash -l -c "pip install awscli"; \
  bash -l -c "pip install s3cmd"; \
