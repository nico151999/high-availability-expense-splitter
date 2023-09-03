package expense

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	expensesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/service"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var streamExpenseIdsAlive = expensesvcv1.StreamExpenseIdsInGroupResponse{
	Update: &expensesvcv1.StreamExpenseIdsInGroupResponse_StillAlive{},
}

func (s *expenseServer) StreamExpenseIdsInGroup(ctx context.Context, req *connect.Request[expensesvcv1.StreamExpenseIdsInGroupRequest], srv *connect.ServerStream[expensesvcv1.StreamExpenseIdsInGroupResponse]) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	if err := service.StreamResource(ctx, s.natsClient, fmt.Sprintf("%s.*", environment.GetExpenseSubject(req.Msg.GetGroupId(), "*")), func(ctx context.Context) (*expensesvcv1.StreamExpenseIdsInGroupResponse, error) {
		return sendCurrentExpenseIds(ctx, s.dbClient, req.Msg.GetGroupId())
	}, srv, &streamExpenseIdsAlive); err != nil {
		if eris.Is(err, errSelectExpenseIds) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBSelectErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSubscribeResource) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed subscribing to updates",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetMessageSubscriptionErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendCurrentResourceMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed returning current resource",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetSendCurrentResourceErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendStreamAliveMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed sending alive message to client",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetSendStreamAliveErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else {
			return connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return nil
}

func sendCurrentExpenseIds(ctx context.Context, dbClient bun.IDB, groupId string) (*expensesvcv1.StreamExpenseIdsInGroupResponse, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	var expenseIds []string
	if err := dbClient.NewSelect().Model((*model.ExpenseModel)(nil)).Where("group_id = ?", groupId).Column("id").Scan(ctx, &expenseIds); err != nil {
		log.Error("failed getting expense IDs", logging.Error(err))
		// TODO: determine reason why expense IDs couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectExpenseIds
	}
	return &expensesvcv1.StreamExpenseIdsInGroupResponse{
		Update: &expensesvcv1.StreamExpenseIdsInGroupResponse_Ids{
			Ids: &expensesvcv1.StreamExpenseIdsInGroupResponse_ExpenseIds{
				Ids: expenseIds,
			},
		},
	}, nil
}
