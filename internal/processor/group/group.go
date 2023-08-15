package group

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
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
func (rpProcessor *groupProcessor) Process(ctx context.Context) (func(ctx context.Context) error, error) {
	var gcSub *nats.Subscription
	{
		eventSubject := environment.GetGroupCreatedSubject("*")
		var err error
		gcSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.groupCreated)
		if err != nil {
			return nil, eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var gdSub *nats.Subscription
	{
		eventSubject := environment.GetGroupDeletedSubject("*")
		var err error
		gdSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.groupDeleted)
		if err != nil {
			return nil, eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var guSub *nats.Subscription
	{
		eventSubject := environment.GetGroupUpdatedSubject("*")
		var err error
		guSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.groupUpdated)
		if err != nil {
			return nil, eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	return processor.GetUnsubscribeSubscriptionsFunc(gcSub, gdSub, guSub), nil
}
