package expensecategoryrelation

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	expensecategoryrelationv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensecategoryrelation/v1"
	expensecategoryrelationprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensecategoryrelation/v1"
	expensecategoryrelationsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensecategoryrelation/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *expensecategoryrelationServer) DeleteExpenseCategoryRelation(ctx context.Context, req *connect.Request[expensecategoryrelationsvcv1.DeleteExpenseCategoryRelationRequest]) (*connect.Response[expensecategoryrelationsvcv1.DeleteExpenseCategoryRelationResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expenseId",
				req.Msg.GetExpenseId()),
			logging.String(
				"categoryId",
				req.Msg.GetCategoryId()),
		),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := deleteExpenseCategoryRelation(ctx, s.natsClient, s.dbClient, req.Msg.GetExpenseId(), req.Msg.GetCategoryId()); err != nil {
		if eris.Is(err, errDeleteExpenseCategoryRelation) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBDeleteErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errNoExpenseCategoryRelationWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the expense category relation ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensecategoryrelationsvcv1.DeleteExpenseCategoryRelationResponse{}), nil
}

func deleteExpenseCategoryRelation(ctx context.Context, nc *nats.EncodedConn, dbClient bun.IDB, expenseId string, categoryId string) error {
	log := logging.FromContext(ctx)

	return dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		expensecategoryrelation := expensecategoryrelationv1.ExpenseCategoryRelation{
			ExpenseId:  expenseId,
			CategoryId: categoryId,
		}
		if _, err := tx.NewDelete().Model(&expensecategoryrelation).WherePK().Exec(ctx); err != nil {
			if eris.Is(err, sql.ErrNoRows) {
				log.Info("expense category relation not found", logging.Error(err))
				return errNoExpenseCategoryRelationWithId
			}
			log.Error("failed deleting expense category relation", logging.Error(err))
			return errDeleteExpenseCategoryRelation
		}
		expense, err := util.CheckResourceExists[*model.Expense](ctx, tx, expensecategoryrelation.GetExpenseId())
		if err != nil {
			return err
		}

		if err := nc.Publish(environment.GetExpenseCategoryRelationDeletedSubject(
			expense.GetGroupId(),
			expenseId,
			categoryId),
			&expensecategoryrelationprocv1.ExpenseCategoryRelationDeleted{
				ExpenseId:  expenseId,
				CategoryId: categoryId,
			}); err != nil {
			log.Error("failed publishing expense category relation deleted event", logging.Error(err))
			return errPublishExpenseCategoryRelationDeleted
		}

		return nil
	})
}
