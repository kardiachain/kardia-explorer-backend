FROM golang:1.13.8 AS protoc_builder

RUN apt-get update && apt-get install -y --no-install-recommends curl make git unzip

FROM protoc_builder AS dev_env
ARG github_token
ENV GO111MODULE on
ENV GOPROXY https://gocenter.io
ENV GOPRIVATE github.com/kardiachain

RUN go get github.com/cortesi/modd/cmd/modd@b0c08baa4ed03a83ba48050774df55f00bef00cb
RUN go get -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.6.2
RUN go get github.com/cucumber/godog/cmd/godog
COPY ./go.mod /
COPY ./go.sum /
RUN go mod download
