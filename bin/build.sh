#!/bin/bash

: ${DISTDIR:='dist'}

cd $(dirname $0)/..

if [ ! -d "${DISTDIR}" ]; then
  mkdir -p dist
fi

version=$(git describe --exact-match --tags HEAD 2> /dev/null)
if [ -z "$version" ]; then
  version=$(git rev-parse --short HEAD)
fi

for os in linux darwin
do
  echo "Building version ${version} for ${os}: ${DISTDIR}/aws-proxy-${version}-${os}-amd64"
  GOOS=${os} GOARCH=amd64 CGO_ENABLED=0 go build -o ${DISTDIR}/aws-proxy-$version-${os}-amd64 aws-proxy.go
done