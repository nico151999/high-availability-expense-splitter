package processor

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"

	"google.golang.org/protobuf/proto"
)

func GetSubjectProcessor[E proto.Message](
	ctx context.Context,
	subjectName string,
	natsClient *nats.EncodedConn,
	processor func(ctx context.Context, event E) error,
) (*nats.Subscription, error) {
	log := logging.FromContext(ctx).With(logging.String("subject", subjectName))
	sub, err := natsClient.Subscribe(subjectName, func(msg E) {
		// if err := msg.InProgress(); err != nil {
		// 	log.Errorw("failed to inform NATS that a message is in progress", logging.Error(err))
		// 	return
		// }
		log.Debug("processing event")
		if err := processor(ctx, msg); err != nil {
			log.Error("failed to process a message", logging.Error(err))
			return
		}
		// if err := msg.Ack(); err != nil {
		// 	log.Errorw("failed to acknowledge a message", logging.Error(err))
		// 	return
		// }
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to subscribe to NATS stream subject '%s'", subjectName)
	}
	return sub, nil
}

func UnsubscribeSubscriptions(ctx context.Context, subs ...*nats.Subscription) error {
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
