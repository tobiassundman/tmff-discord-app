#!/bin/bash

export GOOS=linux
export GOARCH=arm64

go build -o tmff-discord-app cmd/main.go
