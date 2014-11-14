FROM ubuntu:trusty

ENV HOME /root
ENV PATH /usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
ENV GOROOT /usr/local/go

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

RUN \
     wget https://storage.googleapis.com/golang/go1.3.3.linux-amd64.tar.gz  -O - | tar -xzf - -C /usr/local
