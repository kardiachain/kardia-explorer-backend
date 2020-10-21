FROM golang:1.13.9-stretch
RUN apt-get update && apt-get install -y curl make git unzip autoconf automake libtool gettext make g++ texinfo
WORKDIR /root
RUN wget https://github.com/emcrisostomo/fswatch/releases/download/1.14.0/fswatch-1.14.0.tar.gz
RUN tar -xvzf fswatch-1.14.0.tar.gz
WORKDIR /root/fswatch-1.14.0
RUN ./configure
RUN make
RUN make install
ENV GO111MODULE=on
RUN go get github.com/cucumber/godog/cmd/godog@v0.10.0