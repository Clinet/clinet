# Please use govvv when possible! (go install github.com/JoshuaDoes/govvv@latest)
cls

#$env:GO111MODULE="off"
go env -w GOOS=windows GOARCH=amd64 GO111MODULE=on; govvv build -ldflags="-s -w" -o clinet.exe
# go build -ldflags="-s -w" -o clinet.exe
