package expensecategoryrelation

import (
	"context"
	"database/sql"

	expensecategoryrelationv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensecategoryrelation/v1"
	expenseprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	expensecategoryrelationprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensecategoryrelation/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func (rpProcessor *expensecategoryrelationProcessor) expenseDeleted(ctx context.Context, req *expenseprocv1.ExpenseDeleted) error {
	log := logging.FromContext(ctx).With(logging.String("expenseId", req.GetId()))
	log.Info("processing expense.CategoryDeleted event")

	return rpProcessor.dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var expensecategoryrelations []*expensecategoryrelationv1.ExpenseCategoryRelation
		if err := tx.NewDelete().Model(&expensecategoryrelations).Where("expense_id = ?", req.GetId()).Returning("category_id").Scan(ctx); err != nil {
			log.Error("failed deleting expense category relations related to deleted expense", logging.Error(err))
			return errDeleteExpenseCategoryRelations
		}

		g, _ := errgroup.WithContext(ctx)
		for _, c := range expensecategoryrelations {
			expensecategoryrelation := c
			g.Go(func() error {
				marshalled, err := proto.Marshal(&expensecategoryrelationprocv1.ExpenseCategoryRelationDeleted{
					ExpenseId:  expensecategoryrelation.GetExpenseId(),
					CategoryId: req.GetId(),
				})
				if err != nil {
					log.Error("failed marshalling expensecategoryrelation deleted event", logging.Error(err))
					return errMarshalExpenseCategoryRelationDeleted
				}
				if err := rpProcessor.natsClient.Publish(
					environment.GetExpenseCategoryRelationDeletedSubject(
						req.GetGroupId(),
						req.GetId(),
						expensecategoryrelation.GetCategoryId(),
					),
					marshalled,
				); err != nil {
					log.Error("failed publishing expensecategoryrelation deleted event", logging.Error(err))
					return errPublishExpenseCategoryRelationDeleted
				}
				return nil
			})
		}
		return g.Wait()
	})
}
