package expensestake

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1/expensestakev1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqClient "github.com/nico151999/high-availability-expense-splitter/pkg/mq/client"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var _ expensestakev1connect.ExpenseStakeServiceHandler = (*expensestakeServer)(nil)

var errNoExpenseStakeWithId = eris.New("there is no expense stake with that ID")
var errInsertExpenseStake = eris.New("failed inserting expense stake")
var errPublishExpenseStakeCreated = eris.New("failed publishing expense stake created event")
var errPublishExpenseStakeDeleted = eris.New("failed publishing expense stake deleted event")
var errSelectExpenseStakeIds = eris.New("failed selecting expense stake IDs")
var errDeleteExpenseStake = eris.New("failed deleting expense stake")

type expensestakeServer struct {
	dbClient   bun.IDB
	natsClient *nats.EncodedConn
	// TODO: add clients to servers this server will communicate with
}

// NewExpenseStakeServer creates a new instance of expense stake server. The context has no effect on the server's lifecycle.
func NewExpenseStakeServer(ctx context.Context, natsServer, dbUser, dbPass, dbAddr, db string) (*expensestakeServer, error) {
	log := logging.FromContext(ctx).Named("NewExpenseStakeServer")
	ctx = logging.IntoContext(ctx, log)
	return NewExpenseStakeServerWithDBClient(
		ctx,
		client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		natsServer)
}

// NewExpenseStakeServerWithDBClient creates a new instance of expense stake server. The context has no effect on the server's lifecycle.
func NewExpenseStakeServerWithDBClient(ctx context.Context, dbClient bun.IDB, natsServer string) (*expensestakeServer, error) {
	log := logging.FromContext(ctx).Named("NewExpenseStakeServerWithDBClient")
	nc, err := mqClient.NewProtoMQClient(natsServer)
	if err != nil {
		msg := "failed creating NATS client"
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
