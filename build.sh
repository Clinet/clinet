#!/bin/bash
# Please use govvv when possible! (go install github.com/JoshuaDoes/govvv@latest)

export GO111MODULE="on"
export GOOS=linux
export GOARCH=amd64
## govvv build -ldflags="-s -w" -o clinet.app
go build -ldflags="-s -w" -o clinet.app
chmod +x clinet.app
