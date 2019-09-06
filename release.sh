#!/usr/bin/env bash

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/oreka-api main.go

pushd distribution > /dev/null
bash make-deb.sh
popd > /dev/null