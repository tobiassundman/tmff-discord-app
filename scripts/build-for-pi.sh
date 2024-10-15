#!/bin/bash

export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=1

export CC=aarch64-linux-gnu-gcc
export CXX=aarch64-linux-gnu-g++

go build -o tmff-discord-app cmd/main.go
