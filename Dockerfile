FROM kardiachain/explorer-builder
WORKDIR /go/src/github/kardiachain/explorer-backend
ADD . .
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/api
RUN go install
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/grabber
RUN go install
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/backfill
RUN go install
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/verifier
RUN go install
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/receipts
RUN go install
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/watcher
RUN go install
WORKDIR /go/bin
RUN mkdir -p abi
ADD abi /go/bin/abi
