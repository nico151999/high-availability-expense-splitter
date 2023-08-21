package category

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1/categoryv1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var _ categoryv1connect.CategoryServiceHandler = (*categoryServer)(nil)

var errSelectCategory = eris.New("failed selecting category")
var errNoCategoryWithId = eris.New("there is no category with that ID")
var errNoGroupWithId = eris.New("there is no group with that ID")
var errSelectGroup = eris.New("failed selecting group")
var errInsertCategory = eris.New("failed inserting category")
var errMarshalCategoryCreated = eris.New("failed marshalling category created event")
var errPublishCategoryCreated = eris.New("failed publishing category created event")
var errMarshalCategoryDeleted = eris.New("failed marshalling category deleted event")
var errPublishCategoryDeleted = eris.New("failed publishing category deleted event")
var errMarshalCategoryUpdated = eris.New("failed marshalling category updated event")
var errPublishCategoryUpdated = eris.New("failed publishing category updated event")
var errSelectCategoryIds = eris.New("failed selecting category IDs")
var errDeleteCategory = eris.New("failed deleting category")
var errUpdateCategory = eris.New("failed updating category")

type categoryServer struct {
	dbClient   bun.IDB
	natsClient *nats.Conn
	// TODO: add clients to servers this server will communicate with
}

// NewCategoryServer creates a new instance of category server. The context has no effect on the server's lifecycle.
func NewCategoryServer(ctx context.Context, natsServer, dbUser, dbPass, dbAddr, db string) (*categoryServer, error) {
	log := logging.FromContext(ctx).NewNamed("NewCategoryServer")
	ctx = logging.IntoContext(ctx, log)
	return NewCategoryServerWithDBClient(
		ctx,
		client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		natsServer)
}

// NewCategoryServerWithDBClient creates a new instance of category server. The context has no effect on the server's lifecycle.
func NewCategoryServerWithDBClient(ctx context.Context, dbClient bun.IDB, natsServer string) (*categoryServer, error) {
	log := logging.FromContext(ctx).NewNamed("NewCategoryServerWithDBClient")
	nc, err := nats.Connect(natsServer)
	if err != nil {
		msg := "failed connecting to NATS server"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return &categoryServer{
		dbClient:   dbClient,
		natsClient: nc,
	}, nil
}

func (rps *categoryServer) Close() error {
	rps.natsClient.Close()
	return nil
}
