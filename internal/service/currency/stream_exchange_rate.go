package currency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	currencyv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/currency/v1"
	currencysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/currency/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/currency/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type exchangeRate struct {
	exchangeRate  float64
	changeChannel chan float64
	mutex         sync.Mutex
}

// update updates the exchange rate to the new value.
// If it changed it writes the new value to the channel and returns true, otherwise false.
func (er *exchangeRate) update(val float64) bool {
	if val != er.exchangeRate {
		er.mutex.Lock()
		er.exchangeRate = val
		er.mutex.Unlock()
		er.changeChannel <- val
		return true
	}
	return false
}

const tickerPeriod = time.Minute

func (s *currencyServer) StreamExchangeRate(
	ctx context.Context,
	req *connect.Request[currencysvcv1.StreamExchangeRateRequest],
	srv *connect.ServerStream[currencysvcv1.StreamExchangeRateResponse]) error {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"srcCurrency",
				req.Msg.GetSourceCurrencyId()),
			logging.String(
				"destCurrency",
				req.Msg.GetDestinationCurrencyId())))
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	err := streamCurrentExchangeRate(
		ctx,
		srv,
		s.natsClient,
		s.dbClient,
		s.currencyClient,
		req.Msg.GetSourceCurrencyId(),
		req.Msg.GetDestinationCurrencyId())
	if err != nil {
		if eris.Is(err, errCurrencyNoLongerFound) {
			return connect.NewError(
				connect.CodeDataLoss,
				eris.New("the currency does no longer exist"))
		} else if nfErr := new(util.ResourceNotFoundError); eris.As(err, nfErr) {
			return connect.NewError(
				connect.CodeNotFound,
				eris.Errorf("the %s %s does not exist", nfErr.ResourceName, nfErr.ResourceId))
		} else if eris.Is(err, errSubscribeCurrency) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed subscribing to updates",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetMessageSubscriptionErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errSendCurrentExchangeRateMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed returning current resource",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetSendCurrentResourceErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errSendStreamAliveMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed sending alive message to client",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetSendStreamAliveErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else {
			return connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return nil
}

func streamCurrentExchangeRate(
	ctx context.Context,
	srv *connect.ServerStream[currencysvcv1.StreamExchangeRateResponse],
	natsClient *nats.Conn,
	dbClient bun.IDB,
	currencyClient client.Client,
	srcCurrencyId string,
	destCurrencyId string) error {
	log := otel.NewOtelLoggerFromContext(ctx)

	ticker := time.NewTicker(tickerPeriod)
	defer ticker.Stop()

	srcSubject := fmt.Sprintf("%s.*", environment.GetCurrencySubject(srcCurrencyId))
	destSubject := fmt.Sprintf("%s.*", environment.GetCurrencySubject(destCurrencyId))

	curChan := make(chan *nats.Msg)
	for _, subject := range []string{srcSubject, destSubject} {
		s := subject
		sub, err := natsClient.ChanSubscribe(s, curChan)
		if err != nil {
			log.Error("failed subscribing to currency events", logging.Error(err), logging.String("subject", s))
			return errSubscribeCurrency
		}
		defer func() {
			if err := sub.Unsubscribe(); err != nil {
				log.Error("failed unsubscribing from currency events", logging.Error(err), logging.String("subject", s))
			}
		}()
	}

	latestEr := exchangeRate{
		changeChannel: make(chan float64),
	}

	er, err := fetchCurrentExchangeRate(ctx, dbClient, currencyClient, srcCurrencyId, destCurrencyId)
	if err != nil {
		return err
	}
	latestEr.update(er)

	if err := sendCurrentExchangeRate(ctx, srv, er); err != nil {
		return err
	}

loop:
	for {
		select {
		case <-curChan:
			er, err := fetchCurrentExchangeRate(ctx, dbClient, currencyClient, srcCurrencyId, destCurrencyId)
			if err != nil {
				if eris.As(err, &util.ResourceNotFoundError{}) {
					return eris.Wrap(errCurrencyNoLongerFound, err.Error())
				}
				return err
			}
			latestEr.update(er)
		case er := <-latestEr.changeChannel:
			if err := sendCurrentExchangeRate(ctx, srv, er); err != nil {
				return err
			}
			ticker.Reset(tickerPeriod)
		case <-ticker.C:
			er, err := fetchCurrentExchangeRate(ctx, dbClient, currencyClient, srcCurrencyId, destCurrencyId)
			if err != nil {
				if eris.As(err, &util.ResourceNotFoundError{}) {
					return eris.Wrap(errCurrencyNoLongerFound, err.Error())
				}
				return err
			}
			if !latestEr.update(er) {
				if err := srv.Send(&currencysvcv1.StreamExchangeRateResponse{
					Update: &currencysvcv1.StreamExchangeRateResponse_StillAlive{},
				}); err != nil {
					log.Error("failed sending still alive message to client", logging.Error(err))
					return errSendStreamAliveMessage
				}
			}
		case <-ctx.Done():
			log.Info("the context is done")
			break loop
		}
	}
	log.Info("the stream ends now")
	return nil
}

func sendCurrentExchangeRate(
	ctx context.Context,
	srv *connect.ServerStream[currencysvcv1.StreamExchangeRateResponse],
	exchangeRate float64) error {
	log := otel.NewOtelLoggerFromContext(ctx)

	if err := srv.Send(&currencysvcv1.StreamExchangeRateResponse{
		Update: &currencysvcv1.StreamExchangeRateResponse_Rate{
			Rate: exchangeRate,
		},
	}); err != nil {
		log.Error("failed sending current exchange rate to client", logging.Error(err))
		return errSendCurrentExchangeRateMessage
	}

	return nil
}

func fetchCurrentExchangeRate(
	ctx context.Context,
	dbClient bun.IDB,
	currencyClient client.Client,
	srcCurrencyId string,
	destCurrencyId string) (float64, error) {
	srcCurrency, err := util.CheckResourceExists[*currencyv1.Currency](ctx, dbClient, srcCurrencyId)
	if err != nil {
		return 0, err
	}

	destCurrency, err := util.CheckResourceExists[*currencyv1.Currency](ctx, dbClient, destCurrencyId)
	if err != nil {
		return 0, err
	}

	exchangeRate, err := currencyClient.GetLatestExchangeRate(ctx, srcCurrency.GetAcronym(), destCurrency.GetAcronym())
	if err != nil {
		return 0, err
	}

	return exchangeRate, nil
}
