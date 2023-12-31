package service

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
)

type retrieveCurrentResourceFunc[T any] func(context.Context) (*T, error)

const tickerPeriod = time.Minute

var ErrSubscribeResource = eris.New("failed subscribing resource")
var ErrSendStreamAliveMessage = eris.New("failed sending stream alive message")
var ErrSendCurrentResourceMessage = eris.New("failed sending current resource message to client")
var ErrResourceNoLongerFound = eris.New("the resource was no longer found")

func StreamResource[T any](
	ctx context.Context,
	natsClient *nats.Conn,
	subj string,
	retrieveCurrentResource retrieveCurrentResourceFunc[T],
	srv *connect.ServerStream[T],
	stillAliveMsg *T) error {
	log := logging.FromContext(ctx)

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

	if err := sendCurrentResource(ctx, srv, retrieveCurrentResource); err != nil {
		return err
	}

loop:
	for {
		select {
		case <-resChan:
			if err := sendCurrentResource(ctx, srv, retrieveCurrentResource); err != nil {
				if eris.As(err, &util.ResourceNotFoundError{}) {
					return eris.Wrap(ErrResourceNoLongerFound, err.Error())
				}
				return err
			}
			ticker.Reset(tickerPeriod)
		case <-ticker.C:
			if err := srv.Send(stillAliveMsg); err != nil {
				log.Error("failed sending still alive message to client", logging.Error(err))
				return ErrSendStreamAliveMessage
			}
		case <-ctx.Done():
			log.Info("the context is done")
			break loop
		}
	}
	log.Info("the stream ends now")
	return nil
}

func sendCurrentResource[T any](
	ctx context.Context,
	srv *connect.ServerStream[T],
	retrieveCurrentResource retrieveCurrentResourceFunc[T]) error {
	log := logging.FromContext(ctx)

	res, err := retrieveCurrentResource(ctx)
	if err != nil {
		return eris.Wrap(err, "failed to retrieve current resource")
	}

	if err := srv.Send(res); err != nil {
		log.Error("failed sending current resource message to client", logging.Error(err))
		return ErrSendCurrentResourceMessage
	}

	return nil
}
