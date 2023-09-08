package testing

import (
	"context"
	"net"
	"os"
	"testing"

	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1/expensev1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/expense"
	clienttesting "github.com/nico151999/high-availability-expense-splitter/pkg/connect/client/testing"
	servertesting "github.com/nico151999/high-availability-expense-splitter/pkg/connect/server/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
)

// SetupExpenseTest creates gRPC server and client and returns instances of interfaces allowing to close both the server and the client. The passed context has no effect on the server's lifecycle.
func SetupExpenseTest(t *testing.T, ctx context.Context, db bun.IDB) (expensev1connect.ExpenseServiceClient, net.Listener, func() error) {
	log := logging.FromContext(ctx).NewNamed("setupExpenseTest")
	ctx = logging.IntoContext(ctx, log)

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

	ln, shutdownServer := servertesting.StartTestServer(
		t,
		ctx,
		db,
		expense.NewExpenseServerWithDBClient,
		expensev1.RegisterExpenseServiceHandler,
		expensev1connect.NewExpenseServiceHandler)
	cl := clienttesting.SetupTestClient(ln, expensev1connect.NewExpenseServiceClient)
	return cl, ln, shutdownServer
}
