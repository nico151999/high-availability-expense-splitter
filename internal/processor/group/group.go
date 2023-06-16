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
	var gcrSub *nats.Subscription
	{
		eventSubject := environment.GroupCreationRequested
		var err error
		gcrSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.groupCreationRequested)
		if err != nil {
			return nil, eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	// TODO: process the other events as well...
	return processor.GetUnsubscribeSubscriptionsFunc(gcrSub), nil
}
