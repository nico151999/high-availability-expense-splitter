package expensecategoryrelation

import (
	"context"

	expensecategoryrelationv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensecategoryrelation/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *expensecategoryrelationProcessor) expensecategoryrelationDeleted(ctx context.Context, req *expensecategoryrelationv1.ExpenseCategoryRelationDeleted) error {
	log := logging.FromContext(ctx)
	log.Info("processing expensecategoryrelation.ExpenseCategoryRelationDeleted event",
		logging.String("expenseId", req.GetExpenseId()),
		logging.String("categoryId", req.GetCategoryId()))
	// TODO: actually process message like sending a project deleted notification and publish an event telling what was done (e.g. project deleted notification sent)
	return nil
}
