#!/bin/bash

set -e
cd $(dirname $(readlink -f $0))

case $1 in
  linux)
    rm -f mtexp
    CGO_ENABLED=1 CC=x86_64-linux-gnu-gcc GOOS=linux GOARCH=amd64 go build
  ;;
  windows)
    rm -f mtexp.exe
    CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build
  ;;
  macos)
    rm -f mtexp 
    CGO_ENABLED=1 CC=x86_64-apple-darwin15-cc GOOS=darwin GOARCH=amd64 go build
  ;;
  js)
    rm -f mtexp.js
    gopherjs build -m -o mtexp.js
  ;;
esac
