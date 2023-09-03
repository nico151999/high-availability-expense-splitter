package testing

// this package contains helpers for testing the cusproj package

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1/groupv1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/group"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqTesting "github.com/nico151999/high-availability-expense-splitter/pkg/mq/testing"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/test/bufconn"
)

func SetupGroupTestClient(t *testing.T, ln *bufconn.Listener) groupv1connect.GroupServiceClient {
	catClient := groupv1connect.NewGroupServiceClient(
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
	return catClient
}

// StartGroupTestServer starts a test group server and returns a listener as well as a function allowing it to be closed. The passed context has no effect on the server's lifecycle.
func StartGroupTestServer(t *testing.T, ctx context.Context, dbClient bun.IDB) (*bufconn.Listener, func() error) {
	log := logging.FromContext(ctx).NewNamed("StartGroupTestServer")
	ctx = logging.IntoContext(ctx, log)

	ln := bufconn.Listen(1024 * 1024)

	for k, v := range map[string]string{
		"K8S_GET_REQUEST_ERROR_REASON": "K8S_GET_REQUEST_ERROR",
		"GLOBAL_DOMAIN":                "de.test",
		"DB_SELECT_ERROR_REASON":       "DB_SELECT_ERROR",
		"DB_DELETE_ERROR_REASON":       "DB_DELETE_ERROR",
		"DB_UPDATE_ERROR_REASON":       "DB_UPDATE_ERROR",
		"DB_INSERT_ERROR_REASON":       "DB_INSERT_ERROR",
	} {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("failed to set env variable %s: %+v", k, err)
		}
	}

	s, natsPort := mqTesting.RunMQServer(t)

	groupServer, err := group.NewGroupServerWithDBClient(ctx, dbClient, fmt.Sprintf("nats://127.0.0.1:%d", natsPort))
	if err != nil {
		t.Fatal("failed to create group server", err)
	}

	connectServer, err := server.NewServer[groupv1connect.GroupServiceHandler](
		ctx,
		bufconn.Listen(1024*1024),
		groupServer,
		groupv1.RegisterGroupServiceHandler,
		groupv1connect.NewGroupServiceHandler,
		"TestGroupService",
		tracetest.NewInMemoryExporter(),
	)
	if err != nil {
		t.Fatal("failed to create grpc server", err)
	}

	go func() {
		if err := connectServer.Serve(
			ctx,
			ln,
		); err != nil {
			t.Error("server exited with error", err)
		}
	}()

	return ln, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		errGroup, ctx := errgroup.WithContext(ctx)
		errGroup.Go(func() error {
			return connectServer.Shutdown(ctx)
		})
		errGroup.Go(groupServer.Close)
		s.Shutdown()
		return errGroup.Wait()
	}
}

// SetupGroupTest creates gRPC server and client and returns instances of interfaces allowing to close both the server and the client. The passed context has no effect on the server's lifecycle.
func SetupGroupTest(t *testing.T, ctx context.Context, db bun.IDB) (groupv1connect.GroupServiceClient, net.Listener, func() error) {
	log := logging.FromContext(ctx).NewNamed("setupGroupTest")
	ctx = logging.IntoContext(ctx, log)

	ln, shutdownServer := StartGroupTestServer(t, ctx, db)
	cl := SetupGroupTestClient(t, ln)
	return cl, ln, shutdownServer
}
