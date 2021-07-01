module chainmaker.org/chainmaker-contract-sdk-docker-go/shim

go 1.15

require (
	chainmaker.org/chainmaker-contract-sdk-docker-go/logger v0.0.0-00010101000000-000000000000
	chainmaker.org/chainmaker-contract-sdk-docker-go/pb v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.5.2
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.38.0
)

replace (
	chainmaker.org/chainmaker-contract-sdk-docker-go/logger => ../logger
	chainmaker.org/chainmaker-contract-sdk-docker-go/pb => ../pb
)
