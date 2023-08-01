package service

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
)

var ErrSubscribeResource = eris.New("failed subscribing resource")
var ErrSendStreamAliveMessage = eris.New("failed sending stream alive message")

func StreamResource[T any](
	ctx context.Context,
	natsClient *nats.Conn,
	subj string,
	sendCurrentResource func(ctx context.Context, srv *connect.ServerStream[T]) error,
	srv *connect.ServerStream[T],
	stillAliveMsg *T) error {
	log := otel.NewOtelLoggerFromContext(ctx)

	const tickerPeriod = time.Minute
	ticker := time.NewTicker(tickerPeriod)
	defer ticker.Stop()

	resChan := make(chan *nats.Msg)
	sub, err := natsClient.ChanSubscribe(subj, resChan)
	if err != nil {
		log.Error("failed subscribing to resource events", logging.Error(err), logging.String("subject", subj))
		return ErrSubscribeResource
	}
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			log.Error("failed unsubscribing from resource events", logging.Error(err), logging.String("subject", subj))
		}
	}()

	if err := sendCurrentResource(ctx, srv); err != nil {
		return err
	}

loop:
	for {
		select {
		case <-resChan:
			if err := sendCurrentResource(ctx, srv); err != nil {
				return err
			}
			ticker.Reset(tickerPeriod)
		case <-ticker.C:
			if err := sendAliveMessage(ctx, srv, stillAliveMsg); err != nil {
				return err
			}
		case <-ctx.Done():
			log.Info("the context is done")
			break loop
		}
	}
	log.Info("the stream ends now")
	return nil
}

func sendAliveMessage[T any](ctx context.Context, srv *connect.ServerStream[T], stillAliveMsg *T) error {
	log := otel.NewOtelLoggerFromContext(ctx)

	if err := srv.Send(stillAliveMsg); err != nil {
		log.Error("failed sending stream alive message to client", logging.Error(err))
		return ErrSendStreamAliveMessage
	}
	return nil
}
