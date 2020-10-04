FROM golang:1.13.8 AS protoc_builder

RUN apt-get update && apt-get install -y --no-install-recommends curl make git unzip autoconf automake libtool gettext make g++ texinfo
WORKDIR /root
RUN wget https://github.com/emcrisostomo/fswatch/releases/download/1.14.0/fswatch-1.14.0.tar.gz
RUN tar -xvzf fswatch-1.14.0.tar.gz
WORKDIR /root/fswatch-1.14.0
RUN ./configure
RUN make
RUN make install

FROM protoc_builder AS dev_env
ARG github_token
ENV GO111MODULE on
ENV GOPROXY https://gocenter.io
ENV GOPRIVATE github.com/kardiachain
RUN git config --global --add url."https://${github_token}:x-oauth-basic@github.com/kardiachain".insteadOf "https://github.com/kardiachain"
RUN go get github.com/cucumber/godog/cmd/godog
COPY ./go.mod /
COPY ./go.sum /
RUN go mod download
