package group

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1/groupv1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqClient "github.com/nico151999/high-availability-expense-splitter/pkg/mq/client"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var _ groupv1connect.GroupServiceHandler = (*groupServer)(nil)

var errNoGroupWithId = eris.New("there is no group with that ID")
var errInsertGroup = eris.New("failed inserting group")
var errPublishGroupCreated = eris.New("failed publishing group created event")
var errPublishGroupDeleted = eris.New("failed publishing group deleted event")
var errPublishGroupUpdated = eris.New("failed publishing group updated event")
var errSelectGroupIds = eris.New("failed selecting group IDs")
var errDeleteGroup = eris.New("failed deleting group")
var errUpdateGroup = eris.New("failed updating group")

type groupServer struct {
	dbClient   bun.IDB
	natsClient *nats.EncodedConn
	// TODO: add clients to servers this server will communicate with
}

// NewGroupServer creates a new instance of group server. The context has no effect on the server's lifecycle.
func NewGroupServer(ctx context.Context, natsServer, dbUser, dbPass, dbAddr, db string) (*groupServer, error) {
	log := logging.FromContext(ctx).Named("NewGroupServer")
	ctx = logging.IntoContext(ctx, log)
	return NewGroupServerWithDBClient(
		ctx,
		client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		natsServer)
}

// NewGroupServerWithDBClient creates a new instance of group server. The context has no effect on the server's lifecycle.
func NewGroupServerWithDBClient(ctx context.Context, dbClient bun.IDB, natsServer string) (*groupServer, error) {
	log := logging.FromContext(ctx).Named("NewGroupServerWithDBClient")
	nc, err := mqClient.NewProtoMQClient(natsServer)
	if err != nil {
		msg := "failed creating NATS client"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return &groupServer{
		dbClient:   dbClient,
		natsClient: nc,
	}, nil
}

func (rps *groupServer) Close() error {
	rps.natsClient.Close()
	return nil
}
