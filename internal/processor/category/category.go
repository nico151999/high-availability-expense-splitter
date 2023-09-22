package category

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/processor"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

type categoryProcessor struct {
	natsClient *nats.Conn
	dbClient   bun.IDB
}

var errDeleteCategories = eris.New("failed deleting categories")
var errMarshalCategoryDeleted = eris.New("could not marshal category deleted message")
var errPublishCategoryDeleted = eris.New("could not publish category deleted event")

// NewCategoryServer creates a new instance of category server.
func NewCategoryProcessor(natsUrl, dbUser, dbPass, dbAddr, db string) (*categoryProcessor, error) {
	nc, err := nats.Connect(natsUrl)
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

	sourceStreamName := environment.GetCategorySourceStreamName()
	groupSourceStreamName := environment.GetGroupSourceStreamName()

	_, err := processor.CreateOrUpdateSourceStream(
		ctx,
		rpProcessor.natsClient,
		sourceStreamName,
		fmt.Sprintf("%s.*", environment.GetCategorySubject("*", "*")),
	)
	if err != nil {
		return err
	}

	var ccCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetCategoryCreatedSubject("*", "*")
		var err error
		ccCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_CATEGORY_PROCESSOR_CATEGORY_CREATED", eventSubject, rpProcessor.categoryCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var cdCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetCategoryDeletedSubject("*", "*")
		var err error
		cdCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_CATEGORY_PROCESSOR_CATEGORY_DELETED", eventSubject, rpProcessor.categoryDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var cuCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetCategoryUpdatedSubject("*", "*")
		var err error
		cuCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_CATEGORY_PROCESSOR_CATEGORY_UPDATED", eventSubject, rpProcessor.categoryUpdated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var gdCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetGroupDeletedSubject("*")
		var err error
		gdCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, groupSourceStreamName, "EXPENSESPLITTER_CATEGORY_PROCESSOR_GROUP_DELETED", eventSubject, rpProcessor.groupDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}

	<-ctx.Done()
	log.Info("the context is done")
	processor.UnsubscribeConsumeContexts(ccCCtx, cdCCtx, cuCCtx, gdCCtx)
	return nil
}
