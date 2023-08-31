package expense

import (
	"context"
	"time"

	"connectrpc.com/connect"
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

func (s *expenseServer) ListExpenseIdsInGroup(ctx context.Context, req *connect.Request[expensesvcv1.ListExpenseIdsInGroupRequest]) (*connect.Response[expensesvcv1.ListExpenseIdsInGroupResponse], error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expenseIds, err := listExpenseIds(ctx, s.dbClient, req.Msg.GetGroupId())
	if err != nil {
		if eris.Is(err, errSelectExpenseIds) {
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

	return connect.NewResponse(&expensesvcv1.ListExpenseIdsInGroupResponse{
		ExpenseIds: expenseIds,
	}), nil
}

func listExpenseIds(ctx context.Context, dbClient bun.IDB, groupId string) ([]string, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	var expenseIds []string
	if err := dbClient.NewSelect().Model((*model.ExpenseModel)(nil)).Where("group_id = ?", groupId).Column("id").Scan(ctx, &expenseIds); err != nil {
		log.Error("failed getting expense IDs", logging.Error(err))
		// TODO: determine reason why expense ID couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectExpenseIds
	}

	return expenseIds, nil
}
