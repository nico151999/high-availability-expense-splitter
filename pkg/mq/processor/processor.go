package processor

import (
	"context"
	"reflect"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"

	"google.golang.org/protobuf/proto"
)

func GetSubjectProcessor[E proto.Message](
	ctx context.Context,
	subjectName string,
	natsClient *nats.Conn,
	processor func(ctx context.Context, event E) error,
) (*nats.Subscription, error) {
	log := logging.FromContext(ctx)
	sub, err := natsClient.Subscribe(subjectName, func(msg *nats.Msg) {
		// if err := msg.InProgress(); err != nil {
		// 	log.Errorw("failed to inform NATS that a message is in progress", logging.String("subject", subjectName), logging.Error(err))
		// 	return
		// }
		var event E
		event = reflect.New(reflect.TypeOf(event).Elem()).Interface().(E)

		if err := proto.Unmarshal(msg.Data, event); err != nil {
			log.Error("failed to unmarshal data of a message", logging.String("subject", subjectName), logging.Error(err))
			return
		}
		if err := processor(ctx, event); err != nil {
			log.Error("failed to process a message", logging.String("subject", subjectName), logging.Error(err))
			return
		}
		// if err := msg.Ack(); err != nil {
		// 	log.Errorw("failed to acknowledge a message", logging.String("subject", subjectName), logging.Error(err))
		// 	return
		// }
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to subscribe to NATS stream subject '%s'", subjectName)
	}
	return sub, nil
}

func GetUnsubscribeSubscriptionsFunc(subs ...*nats.Subscription) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		errGr, _ := errgroup.WithContext(ctx)
		for _, sub := range subs {
			s := sub
			errGr.Go(func() error {
				if err := s.Unsubscribe(); err != nil {
					return eris.Wrapf(err, "failed to unsubscribe subscription on %s event", s.Subject)
				}
				return nil
			})
		}
		return errGr.Wait()
	}
}
