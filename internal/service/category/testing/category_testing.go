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
	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1/categoryv1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/category"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqTesting "github.com/nico151999/high-availability-expense-splitter/pkg/mq/testing"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/test/bufconn"
)

func SetupCategoryTestClient(t *testing.T, ln *bufconn.Listener) categoryv1connect.CategoryServiceClient {
	catClient := categoryv1connect.NewCategoryServiceClient(
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

// StartCategoryTestServer starts a test category server and returns a listener as well as a function allowing it to be closed. The passed context has no effect on the server's lifecycle.
func StartCategoryTestServer(t *testing.T, ctx context.Context, dbClient bun.IDB) (*bufconn.Listener, func() error) {
	log := logging.FromContext(ctx).NewNamed("StartCategoryTestServer")
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

	categoryServer, err := category.NewCategoryServerWithDBClient(ctx, dbClient, fmt.Sprintf("nats://127.0.0.1:%d", natsPort))
	if err != nil {
		t.Fatal("failed to create category server", err)
	}

	connectServer, err := server.NewServer[categoryv1connect.CategoryServiceHandler](
		ctx,
		bufconn.Listen(1024*1024),
		categoryServer,
		categoryv1.RegisterCategoryServiceHandler,
		categoryv1connect.NewCategoryServiceHandler,
		"TestCategoryService",
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
		errCategory, ctx := errgroup.WithContext(ctx)
		errCategory.Go(func() error {
			return connectServer.Shutdown(ctx)
		})
		errCategory.Go(categoryServer.Close)
		s.Shutdown()
		return errCategory.Wait()
	}
}

// SetupCategoryTest creates gRPC server and client and returns instances of interfaces allowing to close both the server and the client. The passed context has no effect on the server's lifecycle.
func SetupCategoryTest(t *testing.T, ctx context.Context, db bun.IDB) (categoryv1connect.CategoryServiceClient, net.Listener, func() error) {
	log := logging.FromContext(ctx).NewNamed("setupCategoryTest")
	ctx = logging.IntoContext(ctx, log)

	ln, shutdownServer := StartCategoryTestServer(t, ctx, db)
	cl := SetupCategoryTestClient(t, ln)
	return cl, ln, shutdownServer
}
