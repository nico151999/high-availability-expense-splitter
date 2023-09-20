package category

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	mqClient "github.com/nico151999/high-availability-expense-splitter/pkg/mq/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/processor"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

type categoryProcessor struct {
	natsClient *nats.EncodedConn
	dbClient   bun.IDB
}

var errDeleteCategories = eris.New("failed deleting categories")
var errMarshalCategoryDeleted = eris.New("could not marshal category deleted message")
var errPublishCategoryDeleted = eris.New("could not publish category deleted event")

// NewCategoryServer creates a new instance of category server.
func NewCategoryProcessor(natsUrl, dbUser, dbPass, dbAddr, db string) (*categoryProcessor, error) {
	nc, err := mqClient.NewProtoMQClient(natsUrl)
	if err != nil {
		return nil, eris.Wrap(err, "failed connecting to NATS server")
	}
	return &categoryProcessor{
		natsClient: nc,
		dbClient:   client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
	}, nil
}

// Process starts the processing of subscriptions and returns a cancel function allowing for cancelation
func (rpProcessor *categoryProcessor) Process(ctx context.Context) error {
	log := logging.FromContext(ctx).Named("Process")
	ctx = logging.IntoContext(ctx, log)

	var ccSub *nats.Subscription
	{
		eventSubject := environment.GetCategoryCreatedSubject("*", "*")
		var err error
		ccSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.categoryCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var cdSub *nats.Subscription
	{
		eventSubject := environment.GetCategoryDeletedSubject("*", "*")
		var err error
		cdSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.categoryDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var cuSub *nats.Subscription
	{
		eventSubject := environment.GetCategoryUpdatedSubject("*", "*")
		var err error
		cuSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.categoryUpdated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var gdSub *nats.Subscription
	{
		eventSubject := environment.GetGroupDeletedSubject("*")
		var err error
		gdSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.groupDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}

	<-ctx.Done()
	log.Info("the context is done")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := processor.UnsubscribeSubscriptions(ctx, ccSub, cdSub, cuSub, gdSub); err != nil {
		return eris.Wrap(err, "failed finalising category processor")
	}
	return nil
}
