package category

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/processor"
	"github.com/rotisserie/eris"
)

type categoryProcessor struct {
	natsClient *nats.Conn
}

// NewCategoryServer creates a new instance of category server.
func NewCategoryProcessor(natsUrl string) (*categoryProcessor, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, eris.Wrap(err, "failed connecting to NATS server")
	}
	return &categoryProcessor{
		natsClient: nc,
	}, nil
}

// Process starts the processing of subscriptions and returns a cancel function allowing for cancelation
func (rpProcessor *categoryProcessor) Process(ctx context.Context) (func(ctx context.Context) error, error) {
	var gcSub *nats.Subscription
	{
		eventSubject := environment.GetCategoryCreatedSubject("*", "*")
		var err error
		gcSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.categoryCreated)
		if err != nil {
			return nil, eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var gdSub *nats.Subscription
	{
		eventSubject := environment.GetCategoryDeletedSubject("*", "*")
		var err error
		gdSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.categoryDeleted)
		if err != nil {
			return nil, eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var guSub *nats.Subscription
	{
		eventSubject := environment.GetCategoryUpdatedSubject("*", "*")
		var err error
		guSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.categoryUpdated)
		if err != nil {
			return nil, eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	return processor.GetUnsubscribeSubscriptionsFunc(gcSub, gdSub, guSub), nil
}
