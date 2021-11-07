#!/bin/sh

export GO111MODULE=on
export PATH=$PATH:~/go/bin

rm -rf build
mkdir -p build
cd build

go get -u golang.org/x/mobile/cmd/gomobile@latest
go get golang.org/x/mobile/bind/objc
gomobile init

gomobile bind -target=ios -v p2p/core