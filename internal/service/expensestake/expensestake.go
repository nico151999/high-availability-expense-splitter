package expensestake

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1/expensestakev1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var _ expensestakev1connect.ExpenseStakeServiceHandler = (*expensestakeServer)(nil)

var errSelectExpenseStake = eris.New("failed selecting expense stake")
var errNoExpenseStakeWithId = eris.New("there is no expense stake with that ID")
var errNoExpenseWithId = eris.New("there is no group with that ID")
var errSelectExpense = eris.New("failed selecting group")
var errNoPersonWithId = eris.New("there is no person with that ID")
var errSelectPerson = eris.New("failed selecting person")
var errInsertExpenseStake = eris.New("failed inserting expense stake")
var errMarshalExpenseStakeCreated = eris.New("failed marshalling expense stake created event")
var errPublishExpenseStakeCreated = eris.New("failed publishing expense stake created event")
var errMarshalExpenseStakeDeleted = eris.New("failed marshalling expense stake deleted event")
var errPublishExpenseStakeDeleted = eris.New("failed publishing expense stake deleted event")
var errMarshalExpenseStakeUpdated = eris.New("failed marshalling expense stake updated event")
var errPublishExpenseStakeUpdated = eris.New("failed publishing expense stake updated event")
var errSelectExpenseStakeIds = eris.New("failed selecting expense stake IDs")
var errDeleteExpenseStake = eris.New("failed deleting expense stake")
var errUpdateExpenseStake = eris.New("failed updating expense stake")

type expensestakeServer struct {
	dbClient   bun.IDB
	natsClient *nats.Conn
	// TODO: add clients to servers this server will communicate with
}

// NewExpenseStakeServer creates a new instance of expense stake server. The context has no effect on the server's lifecycle.
func NewExpenseStakeServer(ctx context.Context, natsServer, dbUser, dbPass, dbAddr, db string) (*expensestakeServer, error) {
	log := logging.FromContext(ctx).NewNamed("NewExpenseStakeServer")
	ctx = logging.IntoContext(ctx, log)
	return NewExpenseStakeServerWithDBClient(
		ctx,
		client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		natsServer)
}

// NewExpenseStakeServerWithDBClient creates a new instance of expense stake server. The context has no effect on the server's lifecycle.
func NewExpenseStakeServerWithDBClient(ctx context.Context, dbClient bun.IDB, natsServer string) (*expensestakeServer, error) {
	log := logging.FromContext(ctx).NewNamed("NewExpenseStakeServerWithDBClient")
	nc, err := nats.Connect(natsServer)
	if err != nil {
		msg := "failed connecting to NATS server"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return &expensestakeServer{
		dbClient:   dbClient,
		natsClient: nc,
	}, nil
}

func (rps *expensestakeServer) Close() error {
	rps.natsClient.Close()
	return nil
}
