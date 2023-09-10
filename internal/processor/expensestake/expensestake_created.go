package expensestake

import (
	"context"

	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensestake/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *expensestakeProcessor) expensestakeCreated(ctx context.Context, req *expensestakev1.ExpenseStakeCreated) error {
	log := logging.FromContext(ctx)
	log.Info("processing expensestake.ExpenseStakeCreated event",
		logging.String("forId", req.GetForId()),
		logging.String("expenseId", req.GetExpenseId()),
		logging.String("expensestakeId", req.GetId()),
		logging.Int32("mainValue", req.GetMainValue()),
		logging.Int32("fractionalValue", req.GetFractionalValue()),
		logging.String("requestorEmail", req.GetRequestorEmail()))
	// TODO: actually process message like sending a project created notification and publish an event telling what was done (e.g. project creation notification sent)
	return nil
}
