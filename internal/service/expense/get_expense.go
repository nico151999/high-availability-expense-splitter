package expense

import (
	"context"
	"time"

	"connectrpc.com/connect"
	expensesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *expenseServer) GetExpense(ctx context.Context, req *connect.Request[expensesvcv1.GetExpenseRequest]) (*connect.Response[expensesvcv1.GetExpenseResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expenseId",
				req.Msg.GetId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expense, err := util.CheckResourceExists[*model.ExpenseModel](ctx, s.dbClient, req.Msg.GetId())
	if err != nil {
		if eris.Is(err, util.ErrSelectResource) {
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
		} else if resErr := new(util.ResourceNotFoundError); eris.As(err, resErr) {
			return nil, connect.NewError(connect.CodeNotFound, eris.Errorf("the %s with ID %s does not exist", resErr.ResourceName, resErr.ResourceId))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensesvcv1.GetExpenseResponse{
		Expense: expense,
	}), nil
}
