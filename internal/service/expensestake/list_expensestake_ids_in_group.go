package expensestake

import (
	"context"
	"time"

	"connectrpc.com/connect"
	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensestake/v1"
	expensestakesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *expensestakeServer) ListExpenseStakeIdsInGroup(ctx context.Context, req *connect.Request[expensestakesvcv1.ListExpenseStakeIdsInGroupRequest]) (*connect.Response[expensestakesvcv1.ListExpenseStakeIdsInGroupResponse], error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expensestakeIds, err := listExpenseStakeIdsInGroup(ctx, s.dbClient, req.Msg.GetGroupId())
	if err != nil {
		if eris.Is(err, errSelectExpenseStakeIds) {
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

	return connect.NewResponse(&expensestakesvcv1.ListExpenseStakeIdsInGroupResponse{
		Ids: expensestakeIds,
	}), nil
}

func listExpenseStakeIdsInGroup(ctx context.Context, dbClient bun.IDB, groupId string) ([]string, error) {
	log := logging.FromContext(ctx)
	var expensestakeIds []string
	if err := dbClient.NewSelect().
		Model((*expensestakev1.ExpenseStake)(nil)).
		Where(
			"expense_id IN (?)",
			dbClient.NewSelect().
				Model((*expensev1.Expense)(nil)).
				Column("id").
				Where("group_id = ?", groupId),
		).
		Column("id").
		Order("for_id ASC").
		Scan(ctx, &expensestakeIds); err != nil {
		log.Error("failed getting expense stake IDs", logging.Error(err))
		// TODO: determine reason why expensestake ID couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectExpenseStakeIds
	}

	return expensestakeIds, nil
}
