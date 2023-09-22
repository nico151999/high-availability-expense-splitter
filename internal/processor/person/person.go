package person

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

type personProcessor struct {
	natsClient *nats.Conn
	dbClient   bun.IDB
}

var errDeletePeople = eris.New("failed deleting people")
var errMarshalPersonDeleted = eris.New("could not marshal person deleted message")
var errPublishPersonDeleted = eris.New("could not publish person deleted event")

// NewPersonServer creates a new instance of person server.
func NewPersonProcessor(natsUrl, dbUser, dbPass, dbAddr, db string) (*personProcessor, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, eris.Wrap(err, "failed connecting to NATS server")
	}
	return &personProcessor{
		natsClient: nc,
		dbClient:   client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
	}, nil
}

// Process starts the processing of subscriptions and returns a cancel function allowing for cancelation
func (rpProcessor *personProcessor) Process(ctx context.Context) error {
	log := logging.FromContext(ctx).Named("Process")
	ctx = logging.IntoContext(ctx, log)

	sourceStreamName := environment.GetPersonSourceStreamName()
	groupSourceStreamName := environment.GetGroupSourceStreamName()

	_, err := processor.CreateOrUpdateSourceStream(
		ctx,
		rpProcessor.natsClient,
		sourceStreamName,
		fmt.Sprintf("%s.*", environment.GetPersonSubject("*", "*")),
	)
	if err != nil {
		return err
	}

	var pcCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetPersonCreatedSubject("*", "*")
		var err error
		pcCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_CATEGORY_PROCESSOR_CATEGORY_CREATED", eventSubject, rpProcessor.personCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var pdCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetPersonDeletedSubject("*", "*")
		var err error
		pdCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_CATEGORY_PROCESSOR_CATEGORY_DELETED", eventSubject, rpProcessor.personDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var puCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetPersonUpdatedSubject("*", "*")
		var err error
		puCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_CATEGORY_PROCESSOR_CATEGORY_UPDATED", eventSubject, rpProcessor.personUpdated)
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
	processor.UnsubscribeConsumeContexts(pcCCtx, pdCCtx, puCCtx, gdCCtx)
	return nil
}
