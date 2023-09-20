package currency

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/currency/v1/currencyv1connect"
	curClient "github.com/nico151999/high-availability-expense-splitter/pkg/currency/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqClient "github.com/nico151999/high-availability-expense-splitter/pkg/mq/client"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var _ currencyv1connect.CurrencyServiceHandler = (*currencyServer)(nil)

var errSelectCurrencies = eris.New("failed selecting currency IDs")
var errSubscribeCurrency = eris.New("failed subscribing to currency subject")
var errSendStreamAliveMessage = eris.New("failed sending stream alive message")
var errSendCurrentExchangeRateMessage = eris.New("failed sending current exchange rate")
var errCurrencyNoLongerFound = eris.New("the currency does no longer exist")

type currencyServer struct {
	dbClient       bun.IDB
	natsClient     *nats.EncodedConn
	currencyClient curClient.Client
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
	nc, err := mqClient.NewProtoMQClient(natsServer)
	if err != nil {
		msg := "failed connecting to NATS server"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return &currencyServer{
		dbClient:       dbClient,
		natsClient:     nc,
		currencyClient: curClient.NewCurrencyClient(),
	}, nil
}

func (rps *currencyServer) Close() error {
	rps.natsClient.Close()
	return nil
}
