.ONESHELL:

build:
	echo "please input contract name, contract name should be same as name in tx: "
	read contract_name
	echo "please input zip file: "
	read zip_file
	go build -o $${contract_name}
	7z a $${zip_file} $${contract_name}

build-mac:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o docker-go-contract-mac-3
