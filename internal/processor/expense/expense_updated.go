package expense

import (
	"context"

	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *expenseProcessor) expenseUpdated(ctx context.Context, req *expensev1.ExpenseUpdated) error {
	log := logging.FromContext(ctx)
	log.Info("processing expense.ExpenseUpdated event",
		logging.String("expenseId", req.GetId()))
	// TODO: actually process message like sending a project updated notification and publish an event telling what was done (e.g. project updated notification sent)
	return nil
}
