package expensestake

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

type expensestakeProcessor struct {
	natsClient *nats.Conn
	dbClient   bun.IDB
}

var errDeleteExpenseStakes = eris.New("failed deleting expense stakes")
var errMarshalExpenseStakeDeleted = eris.New("could not marshal expensestake deleted message")
var errPublishExpenseStakeDeleted = eris.New("could not publish expensestake deleted event")

// NewExpenseStakeServer creates a new instance of expensestake server.
func NewExpenseStakeProcessor(natsUrl, dbUser, dbPass, dbAddr, db string) (*expensestakeProcessor, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, eris.Wrap(err, "failed connecting to NATS server")
	}
	return &expensestakeProcessor{
		natsClient: nc,
		dbClient:   client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
	}, nil
}

// Process starts the processing of subscriptions and returns a cancel function allowing for cancelation
func (rpProcessor *expensestakeProcessor) Process(ctx context.Context) error {
	log := logging.FromContext(ctx).Named("Process")
	ctx = logging.IntoContext(ctx, log)

	sourceStreamName := environment.GetExpenseStakeSourceStreamName()
	expenseSourceStreamName := environment.GetExpenseSourceStreamName()

	_, err := processor.CreateOrUpdateSourceStream(
		ctx,
		rpProcessor.natsClient,
		sourceStreamName,
		fmt.Sprintf("%s.*", environment.GetExpenseStakeSubject("*", "*", "*")),
	)
	if err != nil {
		return err
	}

	var escCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseStakeCreatedSubject("*", "*", "*")
		var err error
		escCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_EXPENSESTAKE_PROCESSOR_EXPENSESTAKE_CREATED", eventSubject, rpProcessor.expensestakeCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var esdCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseStakeDeletedSubject("*", "*", "*")
		var err error
		esdCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_EXPENSESTAKE_PROCESSOR_EXPENSESTAKE_DELETED", eventSubject, rpProcessor.expensestakeDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var esuCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseStakeUpdatedSubject("*", "*", "*")
		var err error
		esuCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_EXPENSESTAKE_PROCESSOR_EXPENSESTAKE_UPDATED", eventSubject, rpProcessor.expensestakeUpdated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var edCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseDeletedSubject("*", "*")
		var err error
		edCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, expenseSourceStreamName, "EXPENSESPLITTER_EXPENSESTAKE_PROCESSOR_EXPENSE_DELETED", eventSubject, rpProcessor.expenseDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}

	<-ctx.Done()
	log.Info("the context is done")
	processor.UnsubscribeConsumeContexts(escCCtx, esdCCtx, esuCCtx, edCCtx)
	return nil
}
