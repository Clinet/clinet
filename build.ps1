# Please use govvv when possible! (go install github.com/JoshuaDoes/govvv@latest)
cls

#$env:GO111MODULE="off"
govvv build -ldflags="-s -w" -o clinet.exe
# go build -ldflags="-s -w" -o clinet.exe
