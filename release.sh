#!/usr/bin/env bash
set -e

export GOOS=linux
export GOARCH=amd64

export CGO_ENABLED=0
#export CXX=x86_64-linux-g++
#export CC=x86_64-linux-gcc

go build -o bin/oreka-api main.go

pushd distribution > /dev/null
bash make-deb.sh
popd > /dev/null