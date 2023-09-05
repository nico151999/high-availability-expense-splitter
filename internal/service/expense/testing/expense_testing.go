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
	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1/expensev1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/expense"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqTesting "github.com/nico151999/high-availability-expense-splitter/pkg/mq/testing"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/test/bufconn"
)

func SetupExpenseTestClient(t *testing.T, ln *bufconn.Listener) expensev1connect.ExpenseServiceClient {
	catClient := expensev1connect.NewExpenseServiceClient(
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

// StartExpenseTestServer starts a test expense server and returns a listener as well as a function allowing it to be closed. The passed context has no effect on the server's lifecycle.
func StartExpenseTestServer(t *testing.T, ctx context.Context, dbClient bun.IDB) (*bufconn.Listener, func() error) {
	log := logging.FromContext(ctx).NewNamed("StartExpenseTestServer")
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

	expenseServer, err := expense.NewExpenseServerWithDBClient(ctx, dbClient, fmt.Sprintf("nats://127.0.0.1:%d", natsPort))
	if err != nil {
		t.Fatal("failed to create expense server", err)
	}

	connectServer, err := server.NewServer[expensev1connect.ExpenseServiceHandler](
		ctx,
		bufconn.Listen(1024*1024),
		expenseServer,
		expensev1.RegisterExpenseServiceHandler,
		expensev1connect.NewExpenseServiceHandler,
		"TestExpenseService",
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
		errExpense, ctx := errgroup.WithContext(ctx)
		errExpense.Go(func() error {
			return connectServer.Shutdown(ctx)
		})
		errExpense.Go(expenseServer.Close)
		s.Shutdown()
		return errExpense.Wait()
	}
}

// SetupExpenseTest creates gRPC server and client and returns instances of interfaces allowing to close both the server and the client. The passed context has no effect on the server's lifecycle.
func SetupExpenseTest(t *testing.T, ctx context.Context, db bun.IDB) (expensev1connect.ExpenseServiceClient, net.Listener, func() error) {
	log := logging.FromContext(ctx).NewNamed("setupExpenseTest")
	ctx = logging.IntoContext(ctx, log)

	ln, shutdownServer := StartExpenseTestServer(t, ctx, db)
	cl := SetupExpenseTestClient(t, ln)
	return cl, ln, shutdownServer
}
