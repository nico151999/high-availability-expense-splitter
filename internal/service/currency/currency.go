package currency

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/currency/v1/currencyv1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var _ currencyv1connect.CurrencyServiceHandler = (*currencyServer)(nil)

var errSelectCurrency = eris.New("failed selecting currency")
var errNoCurrencyWithId = eris.New("there is no currency with that ID")
var errInsertCurrency = eris.New("failed inserting currency")
var errMarshalCurrencyCreated = eris.New("failed marshalling currency created event")
var errPublishCurrencyCreated = eris.New("failed publishing currency created event")
var errMarshalCurrencyDeleted = eris.New("failed marshalling currency deleted event")
var errPublishCurrencyDeleted = eris.New("failed publishing currency deleted event")
var errMarshalCurrencyUpdated = eris.New("failed marshalling currency updated event")
var errPublishCurrencyUpdated = eris.New("failed publishing currency updated event")
var errSelectCurrencyIds = eris.New("failed selecting currency IDs")
var errDeleteCurrency = eris.New("failed deleting currency")
var errUpdateCurrency = eris.New("failed updating currency")

type currencyServer struct {
	dbClient   bun.IDB
	natsClient *nats.Conn
	// TODO: add clients to servers this server will communicate with
}

// NewCurrencyServer creates a new instance of currency server. The context has no effect on the server's lifecycle.
func NewCurrencyServer(ctx context.Context, natsServer, dbUser, dbPass, dbAddr, db string) (*currencyServer, error) {
	log := logging.FromContext(ctx).NewNamed("NewCurrencyServer")
	ctx = logging.IntoContext(ctx, log)
	return NewCurrencyServerWithDBClient(
		ctx,
		client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		natsServer)
}

// NewCurrencyServerWithDBClient creates a new instance of currency server. The context has no effect on the server's lifecycle.
func NewCurrencyServerWithDBClient(ctx context.Context, dbClient bun.IDB, natsServer string) (*currencyServer, error) {
	log := logging.FromContext(ctx).NewNamed("NewCurrencyServerWithDBClient")
	nc, err := nats.Connect(natsServer)
	if err != nil {
		msg := "failed connecting to NATS server"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return &currencyServer{
		dbClient:   dbClient,
		natsClient: nc,
	}, nil
}

func (rps *currencyServer) Close() error {
	rps.natsClient.Close()
	return nil
}
