FROM kardiachain/explorer-builder
WORKDIR /go/src/github/kardiachain/explorer-backend
ADD . .
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/api
RUN go install
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/grabber
RUN go install
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/watcher
RUN go install
WORKDIR /go/bin