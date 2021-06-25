package internal

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"context"
	"google.golang.org/grpc"
	"net"
	"time"
)

const (
	dialTimeout        = 10 * time.Second
	maxRecvMessageSize = 100 * 1024 * 1024 // 100 MiB
	maxSendMessageSize = 100 * 1024 * 1024 // 100 MiB
)

// NewClientConn ...
func NewClientConn(sockAddress string) (*grpc.ClientConn, error) {

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, sockAddress string) (net.Conn, error) {
			unixAddress, err := net.ResolveUnixAddr("unix", sockAddress)
			conn, err := net.DialUnix("unix", nil, unixAddress)
			return conn, err
		}),
		grpc.FailOnNonTempDialError(true),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxRecvMessageSize),
			grpc.MaxCallSendMsgSize(maxSendMessageSize),
		),
	}

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	return grpc.DialContext(ctx, sockAddress, dialOpts...)
}

func NewContractClient(conn *grpc.ClientConn) (protogo.Contract_ConnectClient, error) {
	return protogo.NewContractClient(conn).Connect(context.Background())
}
