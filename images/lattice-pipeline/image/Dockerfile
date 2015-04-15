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
    software-properties-common \
    curl \
    libcurl4-openssl-dev \
    python-dev \
    python-pip \
    wget \
    unzip \
    jq; \
  add-apt-repository ppa:git-core/ppa -y; \
  apt-get update; \
  apt-get -y install git

RUN \
  wget --quiet https://dl.bintray.com/mitchellh/terraform/terraform_0.4.2_linux_amd64.zip && \
  unzip terraform_0.4.2_linux_amd64.zip -d /usr/local/terraform; \
  rm -v terraform_0.4.2_linux_amd64.zip; \
  cd /usr/local/bin && ln -sv /usr/local/terraform/* .

RUN \
  bash -l -c "pip install awscli";

RUN \
    git config --global user.email "cf-diego-edge-eng+gocd@pivotal.io"

RUN \
    git config --global user.name "Lattice Continuous Deployment"

RUN \
     wget https://storage.googleapis.com/golang/go1.4.2.src.tar.gz  -O - | tar -xzf - -C /usr/local && \
     cd /usr/local/go/src && \
     GOOS=linux GOARCH=amd64 CGO_ENABLED=1 ./make.bash --no-clean && \
     GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 ./make.bash --no-clean

RUN \
    mkdir -p ~/go/src; \
    GOBIN=/usr/local/bin GOPATH=~/go go get -v github.com/dajulia3/godep; \
    rm -rf ~/go

RUN \
    echo "Host github.com" > /etc/ssh/ssh_config

RUN \
    echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config
