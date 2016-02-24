#!/bin/sh bash
export GOPATH=$GOPATH:/Users/meitu/go-workspace/src/github.com/mtinf/qshell
go build -o mtshell
CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o mtshell_linux