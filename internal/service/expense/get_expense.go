package expense

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
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

func (s *expenseServer) GetExpense(ctx context.Context, req *connect.Request[expensesvcv1.GetExpenseRequest]) (*connect.Response[expensesvcv1.GetExpenseResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expenseId",
				req.Msg.GetExpenseId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expense, err := getExpense(ctx, s.dbClient, req.Msg.GetExpenseId())
	if err != nil {
		if eris.Is(err, errSelectExpense) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBSelectErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errNoExpenseWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the expense ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensesvcv1.GetExpenseResponse{
		Expense: expense,
	}), nil
}

func getExpense(ctx context.Context, dbClient bun.IDB, expenseId string) (*expensev1.Expense, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	expense := &expensev1.Expense{
		Id: expenseId,
	}
	expenseModel := model.NewExpense(expense)
	if err := dbClient.NewSelect().Model(expenseModel).WherePK().Limit(1).Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			log.Debug("expense not found", logging.Error(err))
			return nil, errNoExpenseWithId
		}
		log.Error("failed getting expense", logging.Error(err))
		return nil, errSelectExpense
	}
	expense = expenseModel.IntoExpense()

	return expense, nil
}
