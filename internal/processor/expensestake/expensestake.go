package expensestake

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
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

	var pcSub *nats.Subscription
	{
		eventSubject := environment.GetExpenseStakeCreatedSubject("*", "*", "*")
		var err error
		pcSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.expensestakeCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var pdSub *nats.Subscription
	{
		eventSubject := environment.GetExpenseStakeDeletedSubject("*", "*", "*")
		var err error
		pdSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.expensestakeDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var puSub *nats.Subscription
	{
		eventSubject := environment.GetExpenseStakeUpdatedSubject("*", "*", "*")
		var err error
		puSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.expensestakeUpdated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var edSub *nats.Subscription
	{
		eventSubject := environment.GetExpenseDeletedSubject("*", "*")
		var err error
		edSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.expenseDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}

	<-ctx.Done()
	log.Info("the context is done")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := processor.UnsubscribeSubscriptions(ctx, pcSub, pdSub, puSub, edSub); err != nil {
		return eris.Wrap(err, "failed finalising expensestake processor")
	}
	return nil
}