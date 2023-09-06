package person

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
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
func (rpProcessor *personProcessor) Process(ctx context.Context) (func(ctx context.Context) error, error) {
	var pcSub *nats.Subscription
	{
		eventSubject := environment.GetPersonCreatedSubject("*", "*")
		var err error
		pcSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.personCreated)
		if err != nil {
			return nil, eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var pdSub *nats.Subscription
	{
		eventSubject := environment.GetPersonDeletedSubject("*", "*")
		var err error
		pdSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.personDeleted)
		if err != nil {
			return nil, eris.Wrapf(err, "an error occurred processing subject %s", eventSubject)
		}
	}
	var puSub *nats.Subscription
	{
		eventSubject := environment.GetPersonUpdatedSubject("*", "*")
		var err error
		puSub, err = processor.GetSubjectProcessor(ctx, eventSubject, rpProcessor.natsClient, rpProcessor.personUpdated)
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
	return processor.GetUnsubscribeSubscriptionsFunc(pcSub, pdSub, puSub, gdSub), nil
}