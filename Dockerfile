FROM ubuntu:14.04.2

RUN apt-get update

RUN DEBIAN_FRONTEND=noninteractive apt-get -y install git mercurial golang

# From the official Golang Dockerfile
#  https://github.com/docker-library/golang/blob/master/1.4/Dockerfile
RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p /go/src/github.com/bitrise-io/cmd-bridge
COPY . /go/src/github.com/bitrise-io/cmd-bridge
WORKDIR /go/src/github.com/bitrise-io/cmd-bridge

RUN go get ./...
RUN go install

CMD cmd-bridge
