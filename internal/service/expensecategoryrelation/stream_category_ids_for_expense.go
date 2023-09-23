package expensecategoryrelation

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	expensecategoryrelationv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensecategoryrelation/v1"
	expensecategoryrelationsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensecategoryrelation/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/service"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var streamCategoryIdsForExpenseAlive = expensecategoryrelationsvcv1.StreamCategoryIdsForExpenseResponse{
	Update: &expensecategoryrelationsvcv1.StreamCategoryIdsForExpenseResponse_StillAlive{},
}

func (s *expensecategoryrelationServer) StreamCategoryIdsForExpense(
	ctx context.Context,
	req *connect.Request[expensecategoryrelationsvcv1.StreamCategoryIdsForExpenseRequest],
	srv *connect.ServerStream[expensecategoryrelationsvcv1.StreamCategoryIdsForExpenseResponse],
) error {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expenseId",
				req.Msg.GetExpenseId()),
		),
	)
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	if err := service.StreamResource(ctx, s.natsClient.Conn, fmt.Sprintf("%s.*", environment.GetExpenseCategoryRelationSubject("*", req.Msg.GetExpenseId(), "*")), func(ctx context.Context) (*expensecategoryrelationsvcv1.StreamCategoryIdsForExpenseResponse, error) {
		return sendCurrentCategoryIdsForExpense(ctx, s.dbClient, req.Msg.GetExpenseId())
	}, srv, &streamCategoryIdsForExpenseAlive); err != nil {
		if eris.Is(err, errSelectCategoryIdsForExpense) {
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

func sendCurrentCategoryIdsForExpense(ctx context.Context, dbClient bun.IDB, expenseId string) (*expensecategoryrelationsvcv1.StreamCategoryIdsForExpenseResponse, error) {
	log := logging.FromContext(ctx)

	var categoryIds []string
	if err := dbClient.NewSelect().Model((*expensecategoryrelationv1.ExpenseCategoryRelation)(nil)).Where("expense_id = ?", expenseId).Column("category_id").Scan(ctx, &categoryIds); err != nil {
		log.Error("failed getting expense stake IDs", logging.Error(err))
		// TODO: determine reason why expensecategoryrelation IDs couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectCategoryIdsForExpense
	}
	return &expensecategoryrelationsvcv1.StreamCategoryIdsForExpenseResponse{
		Update: &expensecategoryrelationsvcv1.StreamCategoryIdsForExpenseResponse_CategoryIds_{
			CategoryIds: &expensecategoryrelationsvcv1.StreamCategoryIdsForExpenseResponse_CategoryIds{
				Ids: categoryIds,
			},
		},
	}, nil
}
