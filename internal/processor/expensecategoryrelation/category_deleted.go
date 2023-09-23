package expensecategoryrelation

import (
	"context"
	"database/sql"

	expensecategoryrelationv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensecategoryrelation/v1"
	categoryprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/category/v1"
	expensecategoryrelationprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensecategoryrelation/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func (rpProcessor *expensecategoryrelationProcessor) categoryDeleted(ctx context.Context, req *categoryprocv1.CategoryDeleted) error {
	log := logging.FromContext(ctx).With(logging.String("categoryId", req.GetId()))
	log.Info("processing expense.CategoryDeleted event")

	return rpProcessor.dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var expensecategoryrelations []*expensecategoryrelationv1.ExpenseCategoryRelation
		if err := tx.NewDelete().Model(&expensecategoryrelations).Where("category_id = ?", req.GetId()).Returning("expense_id").Scan(ctx); err != nil {
			log.Error("failed deleting expense category relations related to deleted category", logging.Error(err))
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
				if err := rpProcessor.natsClient.Publish(environment.GetExpenseCategoryRelationDeletedSubject(req.GetGroupId(), expensecategoryrelation.GetExpenseId(), req.GetId()), marshalled); err != nil {
					log.Error("failed publishing expensecategoryrelation deleted event", logging.Error(err))
					return errPublishExpenseCategoryRelationDeleted
				}
				return nil
			})
		}
		return g.Wait()
	})
}
