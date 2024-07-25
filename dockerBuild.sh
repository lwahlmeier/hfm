#!/bin/bash


GIT_TAG=$(git describe --exact-match --tags HEAD 2>/dev/null)
VERSION="v0.0.0-unknown"

echo "Got tag:\"${GIT_TAG}\""
if [ -z $GIT_TAG ]; then
  GIT_BRANCH=$(git branch | grep \* | cut -d ' ' -f2)
  echo "Got branch:\"${GIT_BRANCH}\""
  if [ "$GIT_BRANCH" == "master" ]; then 
    VERSION="v0.0.0-master"
  fi
else
  VERSION=$GIT_TAG
fi

set -e

echo "---------------------"
echo "Building httpFileMerge    "
echo "---------------------"

docker run --rm -e VERSION=${VERSION} -e GO111MODULE=on -e HOME=/tmp -u $(id -u ${USER}):$(id -g ${USER}) -v "$PWD":/go/build -w /go/build golang:1.22 \
./build.sh

echo ""
echo "---------------------"
echo "Building httpFileMerge Container version: ${VERSION}"
echo "---------------------"

DTAG="lwahlmeier/hfm:${VERSION}"

docker build . -t ${DTAG}

echo "---------------------"
echo "Created Tag ${DTAG}"
echo "---------------------"

