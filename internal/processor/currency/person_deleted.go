package currency

import (
	"context"

	currencyv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/currency/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *currencyProcessor) currencyDeleted(ctx context.Context, req *currencyv1.CurrencyDeleted) error {
	log := logging.FromContext(ctx)
	log.Info("processing currency.CurrencyDeleted event",
		logging.String("currencyId", req.GetId()))
	// TODO: actually process message like sending a project deleted notification and publish an event telling what was done (e.g. project deleted notification sent)
	return nil
}
