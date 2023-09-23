package expensecategoryrelation

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

type expensecategoryrelationProcessor struct {
	natsClient *nats.Conn
	dbClient   bun.IDB
}

var errDeleteExpenseCategoryRelations = eris.New("failed deleting expense category relations")
var errMarshalExpenseCategoryRelationDeleted = eris.New("could not marshal expensecategoryrelation deleted message")
var errPublishExpenseCategoryRelationDeleted = eris.New("could not publish expensecategoryrelation deleted event")

// NewExpenseCategoryRelationServer creates a new instance of expensecategoryrelation server.
func NewExpenseCategoryRelationProcessor(natsUrl, dbUser, dbPass, dbAddr, db string) (*expensecategoryrelationProcessor, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, eris.Wrap(err, "failed connecting to NATS server")
	}
	return &expensecategoryrelationProcessor{
		natsClient: nc,
		dbClient:   client.NewPostgresDBClient(dbUser, dbPass, dbAddr, db),
	}, nil
}

// Process starts the processing of subscriptions and returns a cancel function allowing for cancelation
func (rpProcessor *expensecategoryrelationProcessor) Process(ctx context.Context) error {
	log := logging.FromContext(ctx).Named("Process")
	ctx = logging.IntoContext(ctx, log)

	sourceStreamName := environment.GetExpenseCategoryRelationSourceStreamName()
	expenseSourceStreamName := environment.GetExpenseSourceStreamName()
	categorySourceStreamName := environment.GetCategorySourceStreamName()

	_, err := processor.CreateOrUpdateSourceStream(
		ctx,
		rpProcessor.natsClient,
		sourceStreamName,
		fmt.Sprintf("%s.*", environment.GetExpenseCategoryRelationSubject("*", "*", "*")),
	)
	if err != nil {
		return err
	}

	var escCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseCategoryRelationCreatedSubject("*", "*", "*")
		var err error
		escCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_EXPENSECATEGORYRELATION_PROCESSOR_EXPENSECATEGORYRELATION_CREATED", eventSubject, rpProcessor.expensecategoryrelationCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var esdCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseCategoryRelationDeletedSubject("*", "*", "*")
		var err error
		esdCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_EXPENSECATEGORYRELATION_PROCESSOR_EXPENSECATEGORYRELATION_DELETED", eventSubject, rpProcessor.expensecategoryrelationDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var edCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetExpenseDeletedSubject("*", "*")
		var err error
		edCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, expenseSourceStreamName, "EXPENSESPLITTER_EXPENSECATEGORYRELATION_PROCESSOR_EXPENSE_DELETED", eventSubject, rpProcessor.expenseDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var cdCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetCategoryDeletedSubject("*", "*")
		var err error
		cdCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, categorySourceStreamName, "EXPENSESPLITTER_EXPENSECATEGORYRELATION_PROCESSOR_CATEGORY_DELETED", eventSubject, rpProcessor.categoryDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}

	<-ctx.Done()
	log.Info("the context is done")
	processor.UnsubscribeConsumeContexts(escCCtx, esdCCtx, edCCtx, cdCCtx)
	return nil
}
