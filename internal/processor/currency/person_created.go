package currency

import (
	"context"

	currencyv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/currency/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *currencyProcessor) currencyCreated(ctx context.Context, req *currencyv1.CurrencyCreated) error {
	log := logging.FromContext(ctx)
	log.Info("processing currency.CurrencyCreated event",
		logging.String("name", req.GetName()),
		logging.String("currencyId", req.GetId()),
		logging.String("requestorEmail", req.GetRequestorEmail()))
	// TODO: actually process message like sending a project created notification and publish an event telling what was done (e.g. project creation notification sent)
	return nil
}
