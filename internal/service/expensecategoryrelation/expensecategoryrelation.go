package expensecategoryrelation

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensecategoryrelation/v1/expensecategoryrelationv1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqClient "github.com/nico151999/high-availability-expense-splitter/pkg/mq/client"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var _ expensecategoryrelationv1connect.ExpenseCategoryRelationServiceHandler = (*expensecategoryrelationServer)(nil)

var errNoExpenseCategoryRelationWithId = eris.New("there is no expense stake with that ID")
var errInsertExpenseCategoryRelation = eris.New("failed inserting expense stake")
var errPublishExpenseCategoryRelationCreated = eris.New("failed publishing expense stake created event")
var errPublishExpenseCategoryRelationDeleted = eris.New("failed publishing expense stake deleted event")
var errSelectCategoryIdsForExpense = eris.New("failed selecting category IDs by expense")
var errSelectExpenseIdsForCategory = eris.New("failed selecting expense IDs by category")
var errDeleteExpenseCategoryRelation = eris.New("failed deleting expense stake")
var errExpenseAndCategoryInDifferentGroups = eris.New("cannot create relation between expense and category from different groups")

type expensecategoryrelationServer struct {
	dbClient   bun.IDB
	natsClient *nats.EncodedConn
	// TODO: add clients to servers this server will communicate with
}

// NewExpenseCategoryRelationServer creates a new instance of expense stake server. The context has no effect on the server's lifecycle.
func NewExpenseCategoryRelationServer(ctx context.Context, natsServer, dbUser, dbPass, dbAddr, db string) (*expensecategoryrelationServer, error) {
	log := logging.FromContext(ctx).Named("NewExpenseCategoryRelationServer")
	ctx = logging.IntoContext(ctx, log)
	return NewExpenseCategoryRelationServerWithDBClient(
		ctx,
		client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		natsServer)
}

// NewExpenseCategoryRelationServerWithDBClient creates a new instance of expense stake server. The context has no effect on the server's lifecycle.
func NewExpenseCategoryRelationServerWithDBClient(ctx context.Context, dbClient bun.IDB, natsServer string) (*expensecategoryrelationServer, error) {
	log := logging.FromContext(ctx).Named("NewExpenseCategoryRelationServerWithDBClient")
	nc, err := mqClient.NewProtoMQClient(natsServer)
	if err != nil {
		msg := "failed creating NATS client"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return &expensecategoryrelationServer{
		dbClient:   dbClient,
		natsClient: nc,
	}, nil
}

func (rps *expensecategoryrelationServer) Close() error {
	rps.natsClient.Close()
	return nil
}
