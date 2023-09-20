package expense

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
	expenseprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	expensesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *expenseServer) DeleteExpense(ctx context.Context, req *connect.Request[expensesvcv1.DeleteExpenseRequest]) (*connect.Response[expensesvcv1.DeleteExpenseResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expenseId",
				req.Msg.GetId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := deleteExpense(ctx, s.natsClient, s.dbClient, req.Msg.GetId()); err != nil {
		if eris.Is(err, errDeleteExpense) {
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
		} else if eris.Is(err, errNoExpenseWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the expense ID does not exist"))
		} else {
			logging.FromContext(ctx).Error("Test acmaojaovdijo", logging.Error(err))
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensesvcv1.DeleteExpenseResponse{}), nil
}

func deleteExpense(ctx context.Context, nc *nats.EncodedConn, dbClient bun.IDB, expenseId string) error {
	log := otel.NewOtelLoggerFromContext(ctx)

	return dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		expense := &expensev1.Expense{
			Id: expenseId,
		}
		expenseModel := model.NewExpense(expense)
		if err := tx.NewDelete().Model(expenseModel).WherePK().Returning("group_id").Scan(ctx); err != nil {
			if eris.Is(err, sql.ErrNoRows) {
				log.Info("expense not found", logging.Error(err))
				return errNoExpenseWithId
			}
			log.Error("failed deleting expense", logging.Error(err))
			return errDeleteExpense
		}
		expense = expenseModel.IntoProtoExpense()

		if err := nc.Publish(environment.GetExpenseDeletedSubject(expense.GroupId, expenseId), &expenseprocv1.ExpenseDeleted{
			Id:      expenseId,
			GroupId: expense.GroupId,
		}); err != nil {
			log.Error("failed publishing expense deleted event", logging.Error(err))
			return errPublishExpenseDeleted
		}
		return nil
	})
}
