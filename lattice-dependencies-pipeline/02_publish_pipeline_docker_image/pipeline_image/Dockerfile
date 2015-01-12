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
    git config --global user.name "Lattice Continuous Deployment"

RUN \
     wget https://storage.googleapis.com/golang/go1.4.src.tar.gz  -O - | tar -xzf - -C /usr/local && \
     cd /usr/local/go/src && \
     GOOS=linux GOARCH=amd64 CGO_ENABLED=1 ./make.bash --no-clean && \
     GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 ./make.bash --no-clean


RUN \
    echo "Host github.com" > /etc/ssh/ssh_config

RUN \
    echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config
