package expense

import (
	"context"

	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *expenseProcessor) expenseCreated(ctx context.Context, req *expensev1.ExpenseCreated) error {
	log := logging.FromContext(ctx)
	log.Info("processing expense.ExpenseCreated event",
		logging.String("name", req.GetName()),
		logging.String("expenseId", req.GetId()),
		logging.String("requestorEmail", req.GetRequestorEmail()))
	// TODO: actually process message like sending a project created notification and publish an event telling what was done (e.g. project creation notification sent)
	return nil
}
