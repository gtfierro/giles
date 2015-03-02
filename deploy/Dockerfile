FROM ubuntu

MAINTAINER Gabe Fierro <gtfierro@eecs.berkeley.edu>

RUN apt-get update
RUN apt-get install -y curl git bzr mercurial
RUN curl -O https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz 
RUN tar -C /usr/local/ -xzf go1.4.2.linux-amd64.tar.gz
ENV PATH /usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
ENV GOPATH /go
ENV GOROOT /usr/local/go
RUN go get github.com/gtfierro/giles
WORKDIR /go/src/github.com/gtfierro/giles
ADD . /go/src/github.com/gtfierro/giles
RUN go get
RUN go build
ADD giles.cfg /opt/giles.cfg
ENTRYPOINT ./giles -c /opt/giles.cfg
EXPOSE 8079 2222
