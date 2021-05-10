#!/bin/bash
rm -rf bin
if [[ $(uname -m) == "x86_64" ]]; then
  GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-s -w -X main.version=${VERSION}" -a -o ./bin/hfm
elif [[ $(uname -m) == "aarch64" ]]; then
  GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags="-s -w -X main.version=${VERSION}" -a -o ./bin/hfm
fi

