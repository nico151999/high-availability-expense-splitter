package client

import (
	"context"

	"github.com/nico151999/high-availability-expense-splitter/pkg/closable"
	"github.com/rotisserie/eris"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ closable.Closer = (*Client[any])(nil)

type Client[GRPCCLIENT any] struct {
	client           GRPCCLIENT
	clientConnection *grpc.ClientConn
}

func (c *Client[GRPCCLIENT]) Close(ctx context.Context) error {
	return c.clientConnection.Close()
}

func (c *Client[GRPCCLIENT]) GetGRPCClient() GRPCCLIENT {
	return c.client
}

func NewClient[GRPCCLIENT any](
	grpcServiceClientCreator func(cc grpc.ClientConnInterface) GRPCCLIENT,
	serverAddress string,
	opts ...grpc.DialOption,
) (
	*Client[GRPCCLIENT],
	error,
) {
	dialOptions := append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	clientConn, err := grpc.Dial(
		serverAddress,
		dialOptions...)
	if err != nil {
		return nil, eris.Wrapf(err, "could not connect to remote service '%s'", serverAddress)
	}
	return &Client[GRPCCLIENT]{
		client:           grpcServiceClientCreator(clientConn),
		clientConnection: clientConn,
	}, nil
}
