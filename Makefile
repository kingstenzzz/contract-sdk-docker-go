build-small:
	set GOOS=linux # doesn't work in windows, please set it up manually
	go build -ldflags="-s -w" -o docker-go-contract20_small
	upx --brute docker-go-contract20_small


build-big:
	set GOOS=linux  # doesn't work in windows, please set it up manually
	go build -o docker-go-contract20_big