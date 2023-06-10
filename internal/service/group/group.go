package group

import (
	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1/groupv1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

type groupServer struct {
	groupv1connect.UnimplementedGroupServiceHandler
	dbClient   bun.IDB
	natsClient *nats.Conn
	// TODO: add clients to servers this server will communicate with
}

// NewGroupServer creates a new instance of group server.
func NewGroupServer(natsServer, dbUser, dbPass, dbAddr, db string) (*groupServer, error) {
	return NewGroupServerWithDBClient(
		client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		natsServer)
}

// NewGroupServerWithDBClient creates a new instance of group server.
func NewGroupServerWithDBClient(dbClient bun.IDB, natsServer string) (*groupServer, error) {
	nc, err := nats.Connect(natsServer)
	if err != nil {
		return nil, eris.Wrap(err, "failed connecting to NATS server")
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
