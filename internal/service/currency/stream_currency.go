package currency

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	currencyv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/currency/v1"
	currencysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/currency/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/service"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var streamCurrencyAlive = currencysvcv1.StreamCurrencyResponse{
	Update: &currencysvcv1.StreamCurrencyResponse_StillAlive{},
}

func (s *currencyServer) StreamCurrency(ctx context.Context, req *connect.Request[currencysvcv1.StreamCurrencyRequest], srv *connect.ServerStream[currencysvcv1.StreamCurrencyResponse]) error {
	ctx, cancel := context.WithTimeout(
		logging.IntoContext(
			ctx,
			logging.FromContext(ctx).With(
				logging.String(
					"currencyId",
					req.Msg.GetId()))),
		time.Hour)
	defer cancel()

	streamSubject := fmt.Sprintf("%s.*", environment.GetCurrencySubject(req.Msg.GetId()))
	if err := service.StreamResource(ctx, s.natsClient.Conn, streamSubject, func(ctx context.Context) (*currencysvcv1.StreamCurrencyResponse, error) {
		return sendCurrentCurrency(ctx, s.dbClient, req.Msg.GetId())
	}, srv, &streamCurrencyAlive); err != nil {
		if eris.Is(err, service.ErrResourceNoLongerFound) {
			return connect.NewError(
				connect.CodeDataLoss,
				eris.New("the currency does no longer exist"))
		} else if eris.As(err, &util.ResourceNotFoundError{}) {
			return connect.NewError(
				connect.CodeNotFound,
				eris.New("the currency does not exist"))
		} else if eris.Is(err, service.ErrSubscribeResource) {
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
		} else if eris.Is(err, service.ErrSendCurrentResourceMessage) {
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
		} else if eris.Is(err, service.ErrSendStreamAliveMessage) {
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

func sendCurrentCurrency(ctx context.Context, dbClient bun.IDB, currencyId string) (*currencysvcv1.StreamCurrencyResponse, error) {
	currency, err := util.CheckResourceExists[*currencyv1.Currency](ctx, dbClient, currencyId)
	if err != nil {
		return nil, err
	}
	return &currencysvcv1.StreamCurrencyResponse{
		Update: &currencysvcv1.StreamCurrencyResponse_Currency{
			Currency: currency,
		},
	}, nil
}
