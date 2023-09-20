package person

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1/personv1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqClient "github.com/nico151999/high-availability-expense-splitter/pkg/mq/client"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var _ personv1connect.PersonServiceHandler = (*personServer)(nil)

var errNoPersonWithId = eris.New("there is no person with that ID")
var errInsertPerson = eris.New("failed inserting person")
var errPublishPersonCreated = eris.New("failed publishing person created event")
var errPublishPersonDeleted = eris.New("failed publishing person deleted event")
var errPublishPersonUpdated = eris.New("failed publishing person updated event")
var errSelectPersonIds = eris.New("failed selecting person IDs")
var errDeletePerson = eris.New("failed deleting person")
var errUpdatePerson = eris.New("failed updating person")

type personServer struct {
	dbClient   bun.IDB
	natsClient *nats.EncodedConn
	// TODO: add clients to servers this server will communicate with
}

// NewPersonServer creates a new instance of person server. The context has no effect on the server's lifecycle.
func NewPersonServer(ctx context.Context, natsServer, dbUser, dbPass, dbAddr, db string) (*personServer, error) {
	log := logging.FromContext(ctx).NewNamed("NewPersonServer")
	ctx = logging.IntoContext(ctx, log)
	return NewPersonServerWithDBClient(
		ctx,
		client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		natsServer)
}

// NewPersonServerWithDBClient creates a new instance of person server. The context has no effect on the server's lifecycle.
func NewPersonServerWithDBClient(ctx context.Context, dbClient bun.IDB, natsServer string) (*personServer, error) {
	log := logging.FromContext(ctx).NewNamed("NewPersonServerWithDBClient")
	nc, err := mqClient.NewProtoMQClient(natsServer)
	if err != nil {
		msg := "failed creating NATS client"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return &personServer{
		dbClient:   dbClient,
		natsClient: nc,
	}, nil
}

func (rps *personServer) Close() error {
	rps.natsClient.Close()
	return nil
}
