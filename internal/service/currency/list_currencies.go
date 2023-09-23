package currency

import (
	"context"
	"time"

	"connectrpc.com/connect"
	currencyv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/currency/v1"
	currencysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/currency/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *currencyServer) ListCurrencies(ctx context.Context, req *connect.Request[currencysvcv1.ListCurrenciesRequest]) (*connect.Response[currencysvcv1.ListCurrenciesResponse], error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	currencies, err := listCurrencies(ctx, s.dbClient)
	if err != nil {
		if eris.Is(err, errSelectCurrencies) {
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
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&currencysvcv1.ListCurrenciesResponse{
		Currencies: currencies,
	}), nil
}

func listCurrencies(ctx context.Context, dbClient bun.IDB) ([]*currencyv1.Currency, error) {
	log := logging.FromContext(ctx)
	var currencies []*currencyv1.Currency
	if err := dbClient.NewSelect().Model(&currencies).Order("acronym ASC").Scan(ctx); err != nil {
		log.Error("failed getting currencies", logging.Error(err))
		// TODO: determine reason why currencies couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectCurrencies
	}

	return currencies, nil
}
