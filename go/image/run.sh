#!/usr/bin/env bash

set -ex

GOOS=linux CGO_ENABLED=0 go build -a -ldflags '-w -s' -installsuffix cgo -o bingo/default default.go
GOOS=linux CGO_ENABLED=0 go build -a -ldflags '-w -s' -installsuffix cgo -o bingo/display_output display_output.go
GOOS=linux CGO_ENABLED=0 go build -a -ldflags '-w -s' -installsuffix cgo -o bingo/test test.go
docker build -t local/go_faas:0.0.1 .