# Dockerfile to build qmdserver container images
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
RUN git checkout dev
RUN goop install
WORKDIR /go/src/github.com/pressly/qmd/cmd/qmdserver
RUN ls -al
RUN goop go build
RUN mv qmdserver /usr/local/bin

EXPOSE 8484

CMD /usr/local/bin/qmdserver --disable-auth --user=$USER --pass=$PASS --db-address=$DB --host-nsqd=$HOST --nsqd-addresses=$NSQD --lookupd-addresses=$LOOKUPD --nsqadmin-address=$ADMIN
