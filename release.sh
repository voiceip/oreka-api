#!/usr/bin/env bash

OOS=linux GOARCH=386 CGO_ENABLED=0 go build -o bin/oreka-api main.go
