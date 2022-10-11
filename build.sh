#!/bin/sh
# Please use govvv when possible! (go install github.com/JoshuaDoes/govvv@latest)

govvv build -ldflags="-s -w" -o clinet.app
## go build -ldflags="-s -w" -o clinet.app
chmod +x clinet.app
