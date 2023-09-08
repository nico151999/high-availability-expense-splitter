package expensestake

import (
	"context"
	"database/sql"

	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensestake/v1"
	expenseprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	expensestakeprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensestake/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func (rpProcessor *expensestakeProcessor) expenseDeleted(ctx context.Context, req *expenseprocv1.ExpenseDeleted) error {
	log := logging.FromContext(ctx).With(logging.String("expenseId", req.GetId()))
	log.Info("processing expense.ExpenseDeleted event")

	return rpProcessor.dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		expense, err := util.CheckResourceExists[*expensev1.Expense](ctx, tx, req.GetId())
		if err != nil {
			return err
		}
		var expensestakes []*expensestakev1.ExpenseStake
		if err := tx.NewDelete().Model(&expensestakes).Where("expense_id = ?", req.GetId()).Returning("id").Scan(ctx); err != nil {
			log.Error("failed deleting expense stakes related to deleted expense", logging.Error(err))
			return errDeleteExpenseStakes
		}

		g, _ := errgroup.WithContext(ctx)
		for _, c := range expensestakes {
			expensestake := c
			g.Go(func() error {
				marshalled, err := proto.Marshal(&expensestakeprocv1.ExpenseStakeDeleted{
					Id: expensestake.Id,
				})
				if err != nil {
					log.Error("failed marshalling expensestake deleted event", logging.Error(err))
					return errMarshalExpenseStakeDeleted
				}
				if err := rpProcessor.natsClient.Publish(environment.GetExpenseStakeDeletedSubject(expense.GetGroupId(), req.GetId(), expensestake.Id), marshalled); err != nil {
					log.Error("failed publishing expensestake deleted event", logging.Error(err))
					return errPublishExpenseStakeDeleted
				}
				return nil
			})
		}
		return g.Wait()
	})
}
