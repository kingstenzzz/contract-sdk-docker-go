build-small:
	set GOOS=linux # doesn't work in windows, please set it up manually
	go build -ldflags="-s -w" -o docker-go-contract20_small
	upx --brute docker-go-contract20_small


build-big:
	#set GOOS=linux  # doesn't work in windows, please set it up manually
	go build -o docker-go-contract20_big

build-mac:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o docker-go-contract-mac-3