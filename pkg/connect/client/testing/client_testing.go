package testing

import (
	"context"
	"net"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/grpc/test/bufconn"
)

func SetupTestClient[T any](
	ln *bufconn.Listener,
	construct func(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) T,
) T {
	return construct(
		&http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return ln.DialContext(ctx)
				},
			},
		},
		"http://"+ln.Addr().String(),
		connect.WithGRPC(),
	)
}
