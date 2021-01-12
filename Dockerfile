#-------------------------------------------------------------------------
FROM ubuntu:18.04
MAINTAINER pravin <pravin@duplocloud.net>

###################################################
ARG DOCKER_FOLDER
RUN echo "============= ++ ========== DOCKER_FOLDER ${DOCKER_FOLDER}"
RUN echo ${DOCKER_FOLDER}
##
ENV GOROOT /opt/go
ENV GOPATH /root/.go
ENV DUPLO_HOME=/root/duplo_terraform
##
ENV GOVERSION 1.15.6
ENV TERRAFORM_VERSION=0.14.3
RUN echo "============= ++ ========== GOVERSION ${GOVERSION} TERRAFORM_VERSION ${TERRAFORM_VERSION}"
##
ENV DEBIAN_FRONTEND noninteractive
ENV INITRD No
ENV LANG en_US.UTF-8
##
ENV TF_DEV=true
ENV TF_RELEASE=true
###############################


###############################
RUN  apt-get update && apt-get -y install \
        bash \
        git  \
        make \
        wget \
        vim \
        unzip \
        curl \
        jq \
        ca-certificates \
        openssl

################ MAKE and GO terraform  ###############
RUN cd /opt && wget https://storage.googleapis.com/golang/go${GOVERSION}.linux-amd64.tar.gz && \
    tar zxf go${GOVERSION}.linux-amd64.tar.gz && rm go${GOVERSION}.linux-amd64.tar.gz
RUN mkdir -p $GOPATH
ENV GOBIN="$HOME/go/bin"
RUN mkdir -p $GOBIN
ENV PATH=$PATH:$GOROOT/bin:$GOBIN

WORKDIR $GOPATH/src/github.com/hashicorp/terraform
RUN git clone https://github.com/hashicorp/terraform.git ./ && \
    git checkout v${TERRAFORM_VERSION} && \
    /bin/bash scripts/build.sh
WORKDIR $GOPATH

ADD . /opt/go/src/github.com/duplocloud/terraform-provider-duplocloud
WORKDIR /opt/go/src/github.com/duplocloud/terraform-provider-duplocloud

#RUN go mod init terraform-provider-duplocloud
RUN go mod vendor
RUN make build
RUN make install


### ?? ###
RUN cd /tmp \
  && wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip \
  && unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin

###copy examples
RUN mkdir -p $DUPLO_HOME
WORKDIR  $DUPLO_HOME
ADD ./examples $DUPLO_HOME/examples
COPY ./examples/*.sh $DUPLO_HOME/
#COPY ${DOCKER_FOLDER}/clean_tf.sh $DUPLO_HOME/
RUN chmod +x $DUPLO_HOME/*.sh
WORKDIR /root/duplo_terraform
RUN rm -rf `find -type d -name .terraform`
RUN rm -rf `find -type d -name .terraform.lock.hcl`
RUN rm -rf `find -type d -name terraform.tfstate`
RUN rm -rf `find -type d -name terraform.tfstate.backup`
RUN find . -type f -name '.terraform' -delete
RUN find . -type f -name '*terraform*' -delete
RUN find . -type f -name '*.log*' -delete
##

###clean
RUN rm -rf /tmp/* \
  && rm -rf /var/lib/apt/lists/* \
  rm -rf /var/tmp/*
RUN  apt-get clean \
      && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*


RUN ls -altR /root/.terraform.d
ENTRYPOINT [ "bash"]
#-------------------------------------------------------------------------

