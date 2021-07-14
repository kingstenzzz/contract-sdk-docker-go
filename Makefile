build:
	set GOOS=linux
	go build -ldflags="-s -w" -o docker-go-contract1
	upx --brute docker-go-contract1