package group

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/processor"
	"github.com/rotisserie/eris"
)

type groupProcessor struct {
	natsClient *nats.Conn
}

// NewGroupServer creates a new instance of group server.
func NewGroupProcessor(natsUrl string) (*groupProcessor, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, eris.Wrap(err, "failed connecting to NATS server")
	}
	return &groupProcessor{
		natsClient: nc,
	}, nil
}

// Process starts the processing of subscriptions and returns a cancel function allowing for cancelation
func (rpProcessor *groupProcessor) Process(ctx context.Context) error {
	log := logging.FromContext(ctx).Named("Process")
	ctx = logging.IntoContext(ctx, log)

	sourceStreamName := environment.GetGroupSourceStreamName()

	_, err := processor.CreateOrUpdateSourceStream(
		ctx,
		rpProcessor.natsClient,
		sourceStreamName,
		fmt.Sprintf("%s.*", environment.GetGroupSubject("*")),
	)
	if err != nil {
		return err
	}

	var gcCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetGroupCreatedSubject("*")
		var err error
		gcCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_GROUP_PROCESSOR_GROUP_CREATED", eventSubject, rpProcessor.groupCreated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var gdCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetGroupDeletedSubject("*")
		var err error
		gdCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_GROUP_PROCESSOR_GROUP_DELETED", eventSubject, rpProcessor.groupDeleted)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var guCCtx jetstream.ConsumeContext
	{
		eventSubject := environment.GetGroupUpdatedSubject("*")
		var err error
		guCCtx, err = processor.GetStreamProcessor(ctx, rpProcessor.natsClient, sourceStreamName, "EXPENSESPLITTER_GROUP_PROCESSOR_GROUP_UPDATED", eventSubject, rpProcessor.groupUpdated)
		if err != nil {
			return eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}

	<-ctx.Done()
	log.Info("the context is done")
	processor.UnsubscribeConsumeContexts(gcCCtx, gdCCtx, guCCtx)
	return nil
}
