FROM golang:1.13.8 as dev_env

RUN apt-get update && apt-get install -y --no-install-recommends curl make git unzip

RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.8.0/protoc-3.8.0-linux-x86_64.zip
RUN unzip protoc-3.8.0-linux-x86_64.zip -d protoc3
RUN mv protoc3/bin/* /usr/local/bin/
RUN mv protoc3/include/* /usr/local/include/

RUN git clone --depth 1 https://github.com/gogo/protobuf.git $GOPATH/src/github.com/gogo/protobuf
RUN cd $GOPATH/src/github.com/gogo/protobuf && make install

ENV GO111MODULE on
ENV GOPROXY=https://gocenter.io
ENV GOPRIVATE github.com/kardiachain

RUN go get github.com/cortesi/modd/cmd/modd@b0c08baa4ed03a83ba48050774df55f00bef00cb
RUN go get -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.6.2
RUN go get github.com/cucumber/godog/cmd/godog
RUN go get github.com/mattn/goveralls

ARG github_token
RUN git config --global --add url."https://${github_token}:x-oauth-basic@github.com/kardiachain".insteadOf "https://github.com/kardiachain"
WORKDIR /app/
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download
COPY . .
ENTRYPOINT [ "modd" ]
