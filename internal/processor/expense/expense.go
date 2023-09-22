package expense

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

type expenseProcessor struct {
	natsClient *nats.Conn
	dbClient   bun.IDB
}

var errDeleteExpenses = eris.New("failed deleting expenses")
var errMarshalExpenseDeleted = eris.New("could not marshal expense deleted message")
var errPublishExpenseDeleted = eris.New("could not publish expense deleted event")

// NewExpenseServer creates a new instance of expense server.
func NewExpenseProcessor(natsUrl, dbUser, dbPass, dbAddr, db string) (*expenseProcessor, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, eris.Wrap(err, "failed connecting to NATS server")
	}
	return &expenseProcessor{
		natsClient: nc,
		dbClient:   client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
	}, nil
}

// Process starts the processing of subscriptions and returns a cancel function allowing for cancelation
func (rpProcessor *expenseProcessor) Process(ctx context.Context) error {
	log := logging.FromContext(ctx).Named("Process")
	ctx = logging.IntoContext(ctx, log)

	sourceStreamName := environment.GetExpenseSourceStreamName()
	groupSourceStreamName := environment.GetGroupSourceStreamName()
	personSourceStreamName := environment.GetPersonSourceStreamName()

	_, err := processor.CreateOrUpdateSourceStream(
		ctx,
		rpProcessor.natsClient,
		sourceStreamName,
		fmt.Sprintf("%s.*", environment.GetExpenseSubject("*", "*")),
	)
	if err != nil {
		return err
	}

	var ecCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseCreatedSubject("*", "*")
		var err error
		ecCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_EXPENSE_PROCESSOR_EXPENSE_CREATED", eventSubject, rpProcessor.expenseCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var edCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseDeletedSubject("*", "*")
		var err error
		edCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_EXPENSE_PROCESSOR_EXPENSE_DELETED", eventSubject, rpProcessor.expenseDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var euCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseUpdatedSubject("*", "*")
		var err error
		euCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_EXPENSE_PROCESSOR_EXPENSE_UPDATED", eventSubject, rpProcessor.expenseUpdated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var gdCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetGroupDeletedSubject("*")
		var err error
		gdCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, groupSourceStreamName, "EXPENSESPLITTER_EXPENSE_PROCESSOR_GROUP_DELETED", eventSubject, rpProcessor.groupDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var pdCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetPersonDeletedSubject("*", "*")
		var err error
		pdCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, personSourceStreamName, "EXPENSESPLITTER_EXPENSE_PROCESSOR_PERSON_DELETED", eventSubject, rpProcessor.personDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}

	<-ctx.Done()
	log.Info("the context is done")
	processor.UnsubscribeConsumeContexts(ecCCtx, edCCtx, euCCtx, gdCCtx, pdCCtx)
	return nil
}
