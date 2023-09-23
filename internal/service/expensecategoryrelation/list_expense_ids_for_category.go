package expensecategoryrelation

import (
	"context"
	"time"

	"connectrpc.com/connect"
	expensecategoryrelationv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensecategoryrelation/v1"
	expensecategoryrelationsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensecategoryrelation/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *expensecategoryrelationServer) ListExpenseIdsForCategory(
	ctx context.Context,
	req *connect.Request[expensecategoryrelationsvcv1.ListExpenseIdsForCategoryRequest],
) (*connect.Response[expensecategoryrelationsvcv1.ListExpenseIdsForCategoryResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"categoryId",
				req.Msg.GetCategoryId()),
		),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expenseIds, err := listExpenseIdsForCategory(ctx, s.dbClient, req.Msg.GetCategoryId())
	if err != nil {
		if eris.Is(err, errSelectCategoryIdsForExpense) {
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
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensecategoryrelationsvcv1.ListExpenseIdsForCategoryResponse{
		ExpenseIds: expenseIds,
	}), nil
}

func listExpenseIdsForCategory(ctx context.Context, dbClient bun.IDB, categoryId string) ([]string, error) {
	log := logging.FromContext(ctx)
	var expenseIds []string
	if err := dbClient.NewSelect().Model((*expensecategoryrelationv1.ExpenseCategoryRelation)(nil)).Where("category_id = ?", categoryId).Column("category_id").Scan(ctx, &expenseIds); err != nil {
		log.Error("failed getting expense IDs", logging.Error(err))
		// TODO: determine reason why expenseIds couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectExpenseIdsForCategory
	}

	return expenseIds, nil
}
