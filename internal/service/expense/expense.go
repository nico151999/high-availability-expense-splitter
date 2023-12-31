package expense

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1/expensev1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqClient "github.com/nico151999/high-availability-expense-splitter/pkg/mq/client"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var _ expensev1connect.ExpenseServiceHandler = (*expenseServer)(nil)

var errNoExpenseWithId = eris.New("there is no expense with that ID")
var errInsertExpense = eris.New("failed inserting expense")
var errPublishExpenseCreated = eris.New("failed publishing expense created event")
var errPublishExpenseDeleted = eris.New("failed publishing expense deleted event")
var errPublishExpenseUpdated = eris.New("failed publishing expense updated event")
var errSelectExpenseIds = eris.New("failed selecting expense IDs")
var errDeleteExpense = eris.New("failed deleting expense")
var errUpdateExpense = eris.New("failed updating expense")

type expenseServer struct {
	dbClient   bun.IDB
	natsClient *nats.EncodedConn
	// TODO: add clients to servers this server will communicate with
}

// NewExpenseServer creates a new instance of expense server. The context has no effect on the server's lifecycle.
func NewExpenseServer(ctx context.Context, natsServer, dbUser, dbPass, dbAddr, db string) (*expenseServer, error) {
	log := logging.FromContext(ctx).Named("NewExpenseServer")
	ctx = logging.IntoContext(ctx, log)
	return NewExpenseServerWithDBClient(
		ctx,
		client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		natsServer)
}

// NewExpenseServerWithDBClient creates a new instance of expense server. The context has no effect on the server's lifecycle.
func NewExpenseServerWithDBClient(ctx context.Context, dbClient bun.IDB, natsServer string) (*expenseServer, error) {
	log := logging.FromContext(ctx).Named("NewExpenseServerWithDBClient")
	nc, err := mqClient.NewProtoMQClient(natsServer)
	if err != nil {
		msg := "failed connecting to NATS server"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return &expenseServer{
		dbClient:   dbClient,
		natsClient: nc,
	}, nil
}

func (rps *expenseServer) Close() error {
	rps.natsClient.Close()
	return nil
}
