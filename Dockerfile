FROM kardiachain/explorer-builder
ARG GITHUB_TOKEN
ENV GOPRIVATE github.com/kardiachain
RUN git config --global --add url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/kardiachain".insteadOf "https://github.com/kardiachain"
WORKDIR /go/src/github/kardiachain/explorer-backend
ADD . .
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/api
RUN go install
WORKDIR /go/src/github/kardiachain/explorer-backend/cmd/grabber
RUN go install
WORKDIR /go/bin