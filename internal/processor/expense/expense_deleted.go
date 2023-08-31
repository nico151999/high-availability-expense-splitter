package expense

import (
	"context"

	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *expenseProcessor) expenseDeleted(ctx context.Context, req *expensev1.ExpenseDeleted) error {
	log := logging.FromContext(ctx)
	log.Info("processing expense.ExpenseDeleted event",
		logging.String("expenseId", req.GetExpenseId()))
	// TODO: actually process message like sending a project deleted notification and publish an event telling what was done (e.g. project deleted notification sent)
	return nil
}
