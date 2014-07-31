# Dockerfile to build qmdworker container images
# Based on Ubuntu 14.04

FROM ubuntu:14.04
MAINTAINER David Kua <david.kua@pressly.com>

# Dependencies
RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -qy build-essential curl git bzr mercurial
RUN curl -s https://storage.googleapis.com/golang/go1.3.src.tar.gz | tar -v -C /usr/local -xz
RUN cd /usr/local/go/src && ./make.bash --no-clean 2>&1

# Environment Variables
ENV PATH /usr/local/go/bin:/go/bin:/usr/local/bin:$PATH
ENV GOPATH /go

# Goop
RUN go get github.com/nitrous-io/goop

# QMD
RUN go get github.com/pressly/qmd
WORKDIR /go/src/github.com/pressly/qmd
RUN goop install
WORKDIR /go/src/github.com/pressly/qmd/cmd/qmdworker
RUN goop go build
RUN mv qmdworker /usr/local/bin

CMD /usr/local/bin/qmdworker --script-dir=$SCRIPT_DIR --working-dir=$WORKING_DIR --store-dir=$STORE_DIR --whitelist=$WHITELIST --lookupd-addresses=$LOOKUPD_ADDRESS --host-nsqd=$NSQD
