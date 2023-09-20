package expense

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

type expenseProcessor struct {
	natsClient *nats.EncodedConn
	dbClient   bun.IDB
}

var errDeleteExpenses = eris.New("failed deleting expenses")
var errMarshalExpenseDeleted = eris.New("could not marshal expense deleted message")
var errPublishExpenseDeleted = eris.New("could not publish expense deleted event")

// NewExpenseServer creates a new instance of expense server.
func NewExpenseProcessor(natsUrl, dbUser, dbPass, dbAddr, db string) (*expenseProcessor, error) {
	nc, err := mqClient.NewProtoMQClient(natsUrl)
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

	var ccSub *nats.Subscription
	{
		eventSubject := environment.GetExpenseCreatedSubject("*", "*")
		var err error
		ccSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.expenseCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var cdSub *nats.Subscription
	{
		eventSubject := environment.GetExpenseDeletedSubject("*", "*")
		var err error
		cdSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.expenseDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var cuSub *nats.Subscription
	{
		eventSubject := environment.GetExpenseUpdatedSubject("*", "*")
		var err error
		cuSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.expenseUpdated)
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
	var pdSub *nats.Subscription
	{
		eventSubject := environment.GetPersonDeletedSubject("*", "*")
		var err error
		pdSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.personDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}

	<-ctx.Done()
	log.Info("the context is done")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := processor.UnsubscribeSubscriptions(ctx, ccSub, cdSub, cuSub, gdSub, pdSub); err != nil {
		return eris.Wrap(err, "failed finalising expense processor")
	}
	return nil
}
