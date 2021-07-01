module chainmaker.org/chainmaker-contract-sdk-docker-go

go 1.15

replace (
	chainmaker.org/chainmaker-contract-sdk-docker-go/logger => ./logger
	chainmaker.org/chainmaker-contract-sdk-docker-go/pb => ./pb
	chainmaker.org/chainmaker-contract-sdk-docker-go/shim => ./shim
)

require (
	chainmaker.org/chainmaker-contract-sdk-docker-go/logger v0.0.0-00010101000000-000000000000 // indirect
	chainmaker.org/chainmaker-contract-sdk-docker-go/pb v0.0.0-00010101000000-000000000000
	chainmaker.org/chainmaker-contract-sdk-docker-go/shim v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.18.1 // indirect
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859 // indirect
)
