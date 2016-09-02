#!/bin/sh

cd $(dirname $0)/..
GOPATH=$PWD

version=$(git describe --exact-match --tags HEAD 2> /dev/null)
if [ -z "$version" ]; then
  version=$(git rev-parse --short HEAD)
fi

GOOS=linux GOARCH=amd64 go build -o build/aws-proxy-$version-linux-amd64 aws-proxy
GOOS=darwin GOARCH=amd64 go build -o build/aws-proxy-$version-darwin-amd64 aws-proxy
