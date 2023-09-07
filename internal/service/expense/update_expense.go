package expense

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	expenseprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	expensesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *expenseServer) UpdateExpense(ctx context.Context, req *connect.Request[expensesvcv1.UpdateExpenseRequest]) (*connect.Response[expensesvcv1.UpdateExpenseResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expenseId",
				req.Msg.GetId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expense, err := updateExpense(ctx, s.natsClient, s.dbClient, req.Msg.GetId(), req.Msg.GetUpdateFields())
	if err != nil {
		if eris.Is(err, errUpdateExpense) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBUpdateErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errNoExpenseWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the expense ID does not exist"))
		} else if resErr := new(util.ResourceNotFoundError); eris.As(err, &resErr) {
			return nil, connect.NewError(connect.CodeNotFound, eris.Errorf("the %s with ID %s does not exist", resErr.ResourceName, resErr.ResourceId))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensesvcv1.UpdateExpenseResponse{
		Expense: expense,
	}), nil
}

func updateExpense(ctx context.Context, nc *nats.Conn, dbClient bun.IDB, expenseId string, params []*expensesvcv1.UpdateExpenseRequest_UpdateField) (*expensev1.Expense, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	expense := &expensev1.Expense{
		Id: expenseId,
	}

	if err := dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		query := tx.NewUpdate()
		for _, param := range params {
			switch option := param.GetUpdateOption().(type) {
			case *expensesvcv1.UpdateExpenseRequest_UpdateField_Name:
				expense.Name = &option.Name
				query.Column("name")
			case *expensesvcv1.UpdateExpenseRequest_UpdateField_ById:
				if _, err := util.CheckResourceExists[*personv1.Person](ctx, tx, option.ById); err != nil {
					return err
				}
				expense.ById = option.ById
				query.Column("by_id")
			case *expensesvcv1.UpdateExpenseRequest_UpdateField_Timestamp:
				expense.Timestamp = option.Timestamp
				query.Column("timestamp")
			case *expensesvcv1.UpdateExpenseRequest_UpdateField_CurrencyId:
				// TODO: check if currency exists
				expense.CurrencyId = option.CurrencyId
				query.Column("currency_id")
			}
		}
		expenseModel := model.NewExpense(expense)
		if err := query.Model(expenseModel).WherePK().Returning("group_id").Scan(ctx); err != nil {
			if eris.Is(err, sql.ErrNoRows) {
				log.Info("expense not found", logging.Error(err))
				return errNoExpenseWithId
			}
			log.Error("failed updating expense", logging.Error(err))
			return errUpdateExpense
		}
		expense = expenseModel.IntoExpense()

		marshalled, err := proto.Marshal(&expenseprocv1.ExpenseUpdated{
			Id: expenseId,
		})
		if err != nil {
			log.Error("failed marshalling expense updated event", logging.Error(err))
			return errMarshalExpenseUpdated
		}
		if err := nc.Publish(environment.GetExpenseUpdatedSubject(expense.GroupId, expenseId), marshalled); err != nil {
			log.Error("failed publishing expense updated event", logging.Error(err))
			return errPublishExpenseUpdated
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return expense, nil
}
