package currency

import (
	"context"
	"time"

	currencyv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/currency/v1"
	currencysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/currency/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/currency/client"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"

	"connectrpc.com/connect"
)

func (s *currencyServer) GetExchangeRate(ctx context.Context, req *connect.Request[currencysvcv1.GetExchangeRateRequest]) (*connect.Response[currencysvcv1.GetExchangeRateResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"srcCurrency",
				req.Msg.GetSourceCurrencyId()),
			logging.String(
				"destCurrency",
				req.Msg.GetDestinationCurrencyId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rate, err := getExchangeRate(ctx, s.dbClient, s.currencyClient, req.Msg)
	if err != nil {
		if eris.Is(err, util.ErrSelectResource) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBSelectErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, client.ErrCurrencyExchangeRateNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, eris.New("exchange rate for the specified timestamp does not exist"))
		} else if resErr := new(util.ResourceNotFoundError); eris.As(err, resErr) {
			return nil, connect.NewError(connect.CodeNotFound, eris.Errorf("the %s with ID %s does not exist", resErr.ResourceName, resErr.ResourceId))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&currencysvcv1.GetExchangeRateResponse{
		Rate: rate,
	}), nil
}

func getExchangeRate(ctx context.Context, db bun.IDB, curClient client.Client, msg *currencysvcv1.GetExchangeRateRequest) (float64, error) {
	src, err := util.CheckResourceExists[*currencyv1.Currency](ctx, db, msg.GetSourceCurrencyId())
	if err != nil {
		return 0, err
	}
	dest, err := util.CheckResourceExists[*currencyv1.Currency](ctx, db, msg.GetDestinationCurrencyId())
	if err != nil {
		return 0, err
	}

	exchangeRate, err := curClient.GetExchangeRate(ctx, src.GetAcronym(), dest.GetAcronym(), msg.GetTimestamp().AsTime())
	if err != nil {
		return 0, err
	}

	return exchangeRate, nil
}
