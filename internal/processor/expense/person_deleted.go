package expense

import (
	"context"
	"database/sql"

	expenseprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	personprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func (rpProcessor *expenseProcessor) personDeleted(ctx context.Context, req *personprocv1.PersonDeleted) error {
	log := logging.FromContext(ctx).With(logging.String("personId", req.GetId()))
	log.Info("processing person.PersonDeleted event")

	return rpProcessor.dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var expenseModels []*model.Expense
		if err := tx.NewDelete().Model(&expenseModels).Where("person_id = ?", req.GetId()).Returning("id").Scan(ctx); err != nil {
			log.Error("failed deleting expenses related to deleted person", logging.Error(err))
			return errDeleteExpenses
		}

		g, _ := errgroup.WithContext(ctx)
		for _, e := range expenseModels {
			expense := e
			g.Go(func() error {
				marshalled, err := proto.Marshal(&expenseprocv1.ExpenseDeleted{
					Id: expense.GetId(),
				})
				if err != nil {
					log.Error("failed marshalling expense deleted event", logging.Error(err))
					return errMarshalExpenseDeleted
				}
				if err := rpProcessor.natsClient.Publish(environment.GetExpenseDeletedSubject(req.GetId(), expense.GetId()), marshalled); err != nil {
					log.Error("failed publishing expense deleted event", logging.Error(err))
					return errPublishExpenseDeleted
				}
				return nil
			})
		}
		return g.Wait()
	})
}
