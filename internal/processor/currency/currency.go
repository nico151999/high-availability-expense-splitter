package currency

// TODO: periodically update currency acronyms in database

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	currencyv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/currency/v1"
	currencyprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/currency/v1"
	curClient "github.com/nico151999/high-availability-expense-splitter/pkg/currency/client"
	dbClient "github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqClient "github.com/nico151999/high-availability-expense-splitter/pkg/mq/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/processor"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/proto"
)

type currencyProcessor struct {
	natsClient     *nats.EncodedConn
	dbClient       bun.IDB
	currencyClient curClient.Client
}

const tickerPeriod = time.Hour

var errSelectCurrencyByAcronym = eris.New("failed selecting currency by acronym")
var errInsertNewCurrency = eris.New("failed inserting currency into database")
var errMarshalCurrencyCreated = eris.New("could not marshal currency created message")
var errPublishCurrencyCreated = eris.New("could not publish currency created event")

// NewCurrencyServer creates a new instance of currency server.
func NewCurrencyProcessor(natsUrl, dbUser, dbPass, dbAddr, db string) (*currencyProcessor, error) {
	nc, err := mqClient.NewProtoMQClient(natsUrl)
	if err != nil {
		return nil, eris.Wrap(err, "failed connecting to NATS server")
	}
	return &currencyProcessor{
		natsClient:     nc,
		dbClient:       dbClient.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
		currencyClient: curClient.NewCurrencyClient(),
	}, nil
}

// Process starts the processing of subscriptions and returns a cancel function allowing for cancelation
func (rpProcessor *currencyProcessor) Process(ctx context.Context) error {
	log := logging.FromContext(ctx).Named("Process")
	ctx = logging.IntoContext(ctx, log)

	var ccSub *nats.Subscription
	{
		eventSubject := environment.GetCurrencyCreatedSubject("*")
		var err error
		ccSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.currencyCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var cdSub *nats.Subscription
	{
		eventSubject := environment.GetCurrencyDeletedSubject("*")
		var err error
		cdSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.currencyDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var cuSub *nats.Subscription
	{
		eventSubject := environment.GetCurrencyUpdatedSubject("*")
		var err error
		cuSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.currencyUpdated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}

	if err := rpProcessor.updateCurrencies(ctx); err != nil {
		log.Error("could not update currencies initially", logging.Error(err))
	} else {
		log.Info("successfully updated currencies initially")
	}

	ticker := time.NewTicker(tickerPeriod)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			if err := rpProcessor.updateCurrencies(ctx); err != nil {
				log.Error("could not update currencies", logging.Error(err))
			} else {
				log.Info("successfully updated currencies")
			}
		case <-ctx.Done():
			log.Info("the context is done")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			if err := processor.UnsubscribeSubscriptions(ctx, ccSub, cdSub, cuSub); err != nil {
				return eris.Wrap(err, "failed finalising currency processor")
			}
			break loop
		}
	}

	return nil
}

func (rpProcessor *currencyProcessor) updateCurrencies(ctx context.Context) error {
	log := logging.FromContext(ctx)

	currencies, err := rpProcessor.currencyClient.FetchCurrencies(ctx)
	if err != nil {
		return err
	}

	for acronym, name := range currencies {
		acronym = strings.ToUpper(acronym)
		log := log.With(logging.String("currency", acronym))

		err := rpProcessor.dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			if err := tx.NewSelect().Model((*currencyv1.Currency)(nil)).Where("acronym = ?", acronym).Limit(1).Scan(ctx); err == nil {
				log.Debug("currency already exists in database")
			} else {
				if eris.Is(err, sql.ErrNoRows) {
					log.Info("inserting new currency into database")
					currency := currencyv1.Currency{
						Id:      util.GenerateIdWithPrefix("currency"),
						Acronym: acronym,
						Name:    name,
					}
					if _, err := tx.NewInsert().Model(&currency).Exec(ctx); err != nil {
						log.Error("failed inserting currency into database", logging.Error(err))
						return errInsertNewCurrency
					}

					marshalled, err := proto.Marshal(&currencyprocv1.CurrencyCreated{
						Id:      currency.GetId(),
						Acronym: currency.GetAcronym(),
						Name:    currency.GetName(),
					})
					if err != nil {
						log.Error("failed marshalling currency created event", logging.Error(err))
						return errMarshalCurrencyCreated
					}
					if err := rpProcessor.natsClient.Publish(environment.GetCurrencyCreatedSubject(currency.GetId()), marshalled); err != nil {
						log.Error("failed publishing currency created event", logging.Error(err))
						return errPublishCurrencyCreated
					}
					return nil
				} else {
					log.Error("failed getting currency by acronym", logging.Error(err))
					return errSelectCurrencyByAcronym
				}
			}
			return nil
		})
		if err != nil {
			msg := "failed updating currency"
			log.Error(msg, logging.Error(err))
			return eris.Wrap(err, msg)
		}
	}
	return nil
}
