#!/bin/sh

export GO111MODULE=on
export PATH=$PATH:~/go/bin

rm -rf go.mod go.sum
go mod init p2p
go mod tidy