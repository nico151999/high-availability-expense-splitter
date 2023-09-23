package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqtesting "github.com/nico151999/high-availability-expense-splitter/pkg/mq/testing"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/test/bufconn"
)

// StartTestServer starts a test server and returns a listener as well as a function allowing it to be closed. The passed context has no effect on the server's lifecycle.
func StartTestServer[
	Server any,
	Handler any,
](
	t *testing.T,
	ctx context.Context,
	dbClient bun.IDB,
	construct func(ctx context.Context, dbClient bun.IDB, natsServer string) (Server, error),
	registerServiceHandler server.ServiceHandlerRegistrarFunc,
	createServiceHandler server.ServiceHandlerCreatorFunc[Handler],
) (*bufconn.Listener, func() error) {
	log := logging.FromContext(ctx).Named("StartTestServer")
	ctx = logging.IntoContext(ctx, log)

	ln := bufconn.Listen(1024 * 1024)

	s, natsPort := mqtesting.RunMQServer(-1)

	testServer, err := construct(ctx, dbClient, fmt.Sprintf("nats://127.0.0.1:%d", natsPort))
	if err != nil {
		t.Fatal("failed to create test server", err)
	}
	var temp interface{} = testServer

	connectServer, err := server.NewServer[Handler](
		ctx,
		bufconn.Listen(1024*1024),
		temp.(Handler),
		registerServiceHandler,
		createServiceHandler,
		"TestService",
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
		errTest, ctx := errgroup.WithContext(ctx)
		errTest.Go(func() error {
			return connectServer.Shutdown(ctx)
		})
		if closableServer, ok := temp.(server.ClosableClientsServer); ok {
			errTest.Go(closableServer.Close)
		}
		s.Shutdown()
		return errTest.Wait()
	}
}
