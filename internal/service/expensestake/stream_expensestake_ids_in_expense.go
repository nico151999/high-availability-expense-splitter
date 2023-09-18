package expensestake

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensestake/v1"
	expensestakesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1"
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

var streamExpenseStakeIdsAlive = expensestakesvcv1.StreamExpenseStakeIdsInExpenseResponse{
	Update: &expensestakesvcv1.StreamExpenseStakeIdsInExpenseResponse_StillAlive{},
}

func (s *expensestakeServer) StreamExpenseStakeIdsInExpense(ctx context.Context, req *connect.Request[expensestakesvcv1.StreamExpenseStakeIdsInExpenseRequest], srv *connect.ServerStream[expensestakesvcv1.StreamExpenseStakeIdsInExpenseResponse]) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	if err := service.StreamResource(ctx, s.natsClient, fmt.Sprintf("%s.*", environment.GetExpenseStakeSubject("*", req.Msg.GetExpenseId(), "*")), func(ctx context.Context) (*expensestakesvcv1.StreamExpenseStakeIdsInExpenseResponse, error) {
		return sendCurrentExpenseStakeIds(ctx, s.dbClient, req.Msg.GetExpenseId())
	}, srv, &streamExpenseStakeIdsAlive); err != nil {
		if eris.Is(err, errSelectExpenseStakeIds) {
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

func sendCurrentExpenseStakeIds(ctx context.Context, dbClient bun.IDB, expenseId string) (*expensestakesvcv1.StreamExpenseStakeIdsInExpenseResponse, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	var expensestakeIds []string
	if err := dbClient.NewSelect().Model((*expensestakev1.ExpenseStake)(nil)).Where("expense_id = ?", expenseId).Column("id").Order("for_id ASC").Scan(ctx, &expensestakeIds); err != nil {
		log.Error("failed getting expense stake IDs", logging.Error(err))
		// TODO: determine reason why expensestake IDs couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectExpenseStakeIds
	}
	return &expensestakesvcv1.StreamExpenseStakeIdsInExpenseResponse{
		Update: &expensestakesvcv1.StreamExpenseStakeIdsInExpenseResponse_Ids{
			Ids: &expensestakesvcv1.StreamExpenseStakeIdsInExpenseResponse_ExpenseStakeIds{
				Ids: expensestakeIds,
			},
		},
	}, nil
}
