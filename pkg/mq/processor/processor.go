package processor

import (
	"context"
	"reflect"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"

	"google.golang.org/protobuf/proto"
)

func CreateOrUpdateSourceStream(
	ctx context.Context,
	natsClient *nats.Conn,
	sourceStreamName string,
	subject string,
) (jetstream.Stream, error) {
	log := logging.FromContext(ctx).With(
		logging.String("sourceStream", sourceStreamName),
		logging.String("subject", subject),
	)

	js, err := jetstream.New(natsClient)
	if err != nil {
		msg := "failed creating NATS jetstream client"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	stream, err := createOrUpdateStream(ctx, js, jetstream.StreamConfig{
		Name:      sourceStreamName,
		Subjects:  []string{subject},
		Retention: jetstream.WorkQueuePolicy,
		Discard:   jetstream.DiscardOld,
		Storage:   jetstream.FileStorage,
	})
	if err != nil {
		msg := "failed creating NATS jetstream source stream"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return stream, nil
}

func GetStreamProcessor[E proto.Message](
	ctx context.Context,
	natsClient *nats.Conn,
	sourceStreamName string,
	streamAndConsumerName string,
	subject string,
	processor func(ctx context.Context, event E) error,
) (jetstream.ConsumeContext, error) {
	log := logging.FromContext(ctx).With(
		logging.String("sourceStream", sourceStreamName),
		logging.String("stream", streamAndConsumerName),
		logging.String("consumer", streamAndConsumerName),
		logging.String("subject", subject),
	)

	js, err := jetstream.New(natsClient)
	if err != nil {
		msg := "failed creating NATS jetstream client"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	stream, err := createOrUpdateStream(ctx, js, jetstream.StreamConfig{
		Name:      streamAndConsumerName,
		Retention: jetstream.WorkQueuePolicy,
		Discard:   jetstream.DiscardOld,
		Storage:   jetstream.FileStorage,
		Sources: []*jetstream.StreamSource{
			{
				Name:          sourceStreamName,
				FilterSubject: subject,
			},
		},
	})
	if err != nil {
		msg := "failed creating NATS jetstream stream"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       streamAndConsumerName,
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		ReplayPolicy:  jetstream.ReplayInstantPolicy,
	})
	if err != nil {
		msg := "failed creating NATS jetstream consumer"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}

	consCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		if err := msg.InProgress(); err != nil {
			log.Error("failed to inform NATS that a message is in progress", logging.Error(err))
			return
		}

		var event E
		event = reflect.New(reflect.TypeOf(event).Elem()).Interface().(E)
		if err := proto.Unmarshal(msg.Data(), event); err != nil {
			log.Error("failed to unmarshal data of a message", logging.Error(err))
			return
		}

		log.Debug("processing event")
		if err := processor(ctx, event); err != nil {
			log.Error("failed to process a message", logging.Error(err))
			return
		}

		if err := msg.Ack(); err != nil {
			log.Error("failed to acknowledge a message", logging.Error(err))
			return
		}
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to subscribe to NATS stream subject '%s'", subject)
	}

	return consCtx, nil
}

func UnsubscribeConsumeContexts(cctxs ...jetstream.ConsumeContext) {
	for _, cctx := range cctxs {
		cctx.Stop()
	}
}

// createOrUpdateStream creates a stream if it does not exist or updates it otherwise
// TODO: remove this function once this is merged: https://github.com/nats-io/nats.go/pull/1395
func createOrUpdateStream(ctx context.Context, js jetstream.JetStream, cfg jetstream.StreamConfig) (jetstream.Stream, error) {
	s, err := js.UpdateStream(ctx, cfg)
	if err != nil {
		if !eris.Is(err, jetstream.ErrStreamNotFound) {
			return nil, err
		}
		return js.CreateStream(ctx, cfg)
	}
	return s, nil
}
