package expensecategoryrelation

import (
	"context"

	expensecategoryrelationv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensecategoryrelation/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *expensecategoryrelationProcessor) expensecategoryrelationCreated(ctx context.Context, req *expensecategoryrelationv1.ExpenseCategoryRelationCreated) error {
	log := logging.FromContext(ctx)
	log.Info("processing expensecategoryrelation.ExpenseCategoryRelationCreated event",
		logging.String("expenseId", req.GetExpenseId()),
		logging.String("categoryId", req.GetCategoryId()),
		logging.String("requestorEmail", req.GetRequestorEmail()))
	// TODO: actually process message like sending a project created notification and publish an event telling what was done (e.g. project creation notification sent)
	return nil
}
