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

var streamExpenseidsForCategoryAlive = expensecategoryrelationsvcv1.StreamExpenseIdsForCategoryResponse{
	Update: &expensecategoryrelationsvcv1.StreamExpenseIdsForCategoryResponse_StillAlive{},
}

func (s *expensecategoryrelationServer) StreamExpenseIdsForCategory(
	ctx context.Context,
	req *connect.Request[expensecategoryrelationsvcv1.StreamExpenseIdsForCategoryRequest],
	srv *connect.ServerStream[expensecategoryrelationsvcv1.StreamExpenseIdsForCategoryResponse],
) error {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"categoryId",
				req.Msg.GetCategoryId()),
		),
	)
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	if err := service.StreamResource(
		ctx,
		s.natsClient.Conn,
		fmt.Sprintf("%s.*", environment.GetExpenseCategoryRelationSubject("*", "*", req.Msg.GetCategoryId())),
		func(ctx context.Context) (*expensecategoryrelationsvcv1.StreamExpenseIdsForCategoryResponse, error) {
			return sendCurrentExpenseIdsForCategory(ctx, s.dbClient, req.Msg.GetCategoryId())
		},
		srv,
		&streamExpenseidsForCategoryAlive); err != nil {
		if eris.Is(err, errSelectExpenseIdsForCategory) {
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

func sendCurrentExpenseIdsForCategory(ctx context.Context, dbClient bun.IDB, categoryId string) (*expensecategoryrelationsvcv1.StreamExpenseIdsForCategoryResponse, error) {
	log := logging.FromContext(ctx)

	var expenseIds []string
	if err := dbClient.NewSelect().Model((*expensecategoryrelationv1.ExpenseCategoryRelation)(nil)).Where("category_id = ?", categoryId).Column("expense_id").Scan(ctx, &expenseIds); err != nil {
		log.Error("failed getting expense stake IDs", logging.Error(err))
		// TODO: determine reason why expensecategoryrelation IDs couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectExpenseIdsForCategory
	}
	return &expensecategoryrelationsvcv1.StreamExpenseIdsForCategoryResponse{
		Update: &expensecategoryrelationsvcv1.StreamExpenseIdsForCategoryResponse_ExpenseIds_{
			ExpenseIds: &expensecategoryrelationsvcv1.StreamExpenseIdsForCategoryResponse_ExpenseIds{
				Ids: expenseIds,
			},
		},
	}, nil
}
