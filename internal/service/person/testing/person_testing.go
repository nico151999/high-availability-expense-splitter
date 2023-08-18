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
	natsserver "github.com/nats-io/nats-server/v2/server"
	natstestserver "github.com/nats-io/nats-server/v2/test"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1/personv1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/person"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/test/bufconn"
)

func runNATSServerOnPort(port int) *natsserver.Server {
	opts := natstestserver.DefaultTestOptions
	opts.Port = port
	return natstestserver.RunServer(&opts)
}

func SetupPersonTestClient(t *testing.T, ln *bufconn.Listener) personv1connect.PersonServiceClient {
	catClient := personv1connect.NewPersonServiceClient(
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

// StartPersonTestServer starts a test person server and returns a listener as well as a function allowing it to be closed. The passed context has no effect on the server's lifecycle.
func StartPersonTestServer(t *testing.T, ctx context.Context, dbClient bun.IDB) (*bufconn.Listener, func() error) {
	log := logging.FromContext(ctx).NewNamed("StartPersonTestServer")
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

	natsPort := 6222
	s := runNATSServerOnPort(natsPort)

	personServer, err := person.NewPersonServerWithDBClient(ctx, dbClient, fmt.Sprintf("nats://127.0.0.1:%d", natsPort))
	if err != nil {
		t.Fatal("failed to create person server", err)
	}

	connectServer, err := server.NewServer[personv1connect.PersonServiceHandler](
		ctx,
		bufconn.Listen(1024*1024),
		personServer,
		personv1.RegisterPersonServiceHandler,
		personv1connect.NewPersonServiceHandler,
		"TestPersonService",
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
		errPerson, ctx := errgroup.WithContext(ctx)
		errPerson.Go(func() error {
			return connectServer.Shutdown(ctx)
		})
		errPerson.Go(personServer.Close)
		s.Shutdown()
		return errPerson.Wait()
	}
}

// SetupPersonTest creates gRPC server and client and returns instances of interfaces allowing to close both the server and the client. The passed context has no effect on the server's lifecycle.
func SetupPersonTest(t *testing.T, ctx context.Context, db bun.IDB) (personv1connect.PersonServiceClient, net.Listener, func() error) {
	log := logging.FromContext(ctx).NewNamed("setupPersonTest")
	ctx = logging.IntoContext(ctx, log)

	ln, shutdownServer := StartPersonTestServer(t, ctx, db)
	cl := SetupPersonTestClient(t, ln)
	return cl, ln, shutdownServer
}
