package testing

// this package contains helpers for testing the cusproj package

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	natsserver "github.com/nats-io/gnatsd/server"
	natstestserver "github.com/nats-io/nats-server/test"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1/groupv1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/group"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func runNATSServerOnPort(port int) *natsserver.Server {
	opts := natstestserver.DefaultTestOptions
	opts.Port = port
	return natstestserver.RunServer(&opts)
}

func SetupGroupTestClient(t *testing.T, ln *bufconn.Listener) *client.Client[groupv1.GroupServiceClient] {
	catClient, err := client.NewClient(
		groupv1.NewGroupServiceClient,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return ln.Dial()
		}))
	if err != nil {
		t.Fatalf("Failed dialing bufnet: %+v", err)
	}
	return catClient
}

func StartGroupTestServer(t *testing.T, dbClient bun.IDB) (*bufconn.Listener, *server.Server) {
	ln := bufconn.Listen(1024 * 1024)
	log := logging.GetLogger()
	ctx := logging.IntoContext(context.Background(), log)

	if err := os.Setenv("K8S_GET_REQUEST_ERROR_REASON", "K8S_GET_REQUEST_ERROR"); err != nil {
		t.Fatal("failed to set env variable K8S_GET_REQUEST_ERROR_REASON", err)
	}
	if err := os.Setenv("GLOBAL_DOMAIN", "de.test"); err != nil {
		t.Fatal("failed to set env variable GLOBAL_DOMAIN", err)
	}

	natsPort := 6222
	s := runNATSServerOnPort(natsPort)

	catServer, err := group.NewGroupServerWithDBClient(dbClient, fmt.Sprintf("nats://127.0.0.1:%d", natsPort))
	if err != nil {
		t.Fatal("failed to create group server", err)
	}

	grpcServer, err := server.NewServer[groupv1connect.GroupServiceHandler](
		ctx,
		ln,
		catServer,
		groupv1.RegisterGroupServiceHandler,
		groupv1connect.NewGroupServiceHandler,
		"TestGroupService",
		"test-trace-collector-url", // TODO: mock a trace collector
		[]string{"*"},
		[]string{"*"},
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	)
	if err != nil {
		t.Fatal("failed to create grpc server", err)
	}

	go func() {
		if err := grpcServer.Serve(
			ctx,
			ln,
		); err != nil {
			t.Error("server exited with error", err)
		}
		if err := catServer.Close(); err != nil {
			t.Error("server exited with error", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := grpcServer.Close(ctx); err != nil {
			t.Error(err)
		}
		s.Shutdown()
	}()

	return ln, grpcServer
}
