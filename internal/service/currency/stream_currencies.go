package currency

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	currencyv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/currency/v1"
	currencysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/currency/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/service"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var streamCurrencyIdsAlive = currencysvcv1.StreamCurrenciesResponse{
	Update: &currencysvcv1.StreamCurrenciesResponse_StillAlive{},
}

func (s *currencyServer) StreamCurrencies(ctx context.Context, req *connect.Request[currencysvcv1.StreamCurrenciesRequest], srv *connect.ServerStream[currencysvcv1.StreamCurrenciesResponse]) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	if err := service.StreamResource(ctx, s.natsClient.Conn, fmt.Sprintf("%s.*", environment.GetCurrencySubject("*")), func(ctx context.Context) (*currencysvcv1.StreamCurrenciesResponse, error) {
		return sendCurrentCurrencies(ctx, s.dbClient)
	}, srv, &streamCurrencyIdsAlive); err != nil {
		if eris.Is(err, errSelectCurrencies) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBSelectErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
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

func sendCurrentCurrencies(ctx context.Context, dbClient bun.IDB) (*currencysvcv1.StreamCurrenciesResponse, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	var currencies []*currencyv1.Currency
	if err := dbClient.NewSelect().Model(&currencies).Order("acronym ASC").Scan(ctx); err != nil {
		log.Error("failed getting currencies", logging.Error(err))
		// TODO: determine reason why currencies couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectCurrencies
	}
	return &currencysvcv1.StreamCurrenciesResponse{
		Update: &currencysvcv1.StreamCurrenciesResponse_Currencies_{
			Currencies: &currencysvcv1.StreamCurrenciesResponse_Currencies{
				Currencies: currencies,
			},
		},
	}, nil
}
