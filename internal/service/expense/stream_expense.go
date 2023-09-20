package expense

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	expensesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/service"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var streamExpenseAlive = expensesvcv1.StreamExpenseResponse{
	Update: &expensesvcv1.StreamExpenseResponse_StillAlive{},
}

func (s *expenseServer) StreamExpense(ctx context.Context, req *connect.Request[expensesvcv1.StreamExpenseRequest], srv *connect.ServerStream[expensesvcv1.StreamExpenseResponse]) error {
	ctx, cancel := context.WithTimeout(
		logging.IntoContext(
			ctx,
			logging.FromContext(ctx).With(
				logging.String(
					"expenseId",
					req.Msg.GetId()))),
		time.Hour)
	defer cancel()

	streamSubject := fmt.Sprintf("%s.*", environment.GetExpenseSubject("*", req.Msg.GetId()))
	if err := service.StreamResource(ctx, s.natsClient.Conn, streamSubject, func(ctx context.Context) (*expensesvcv1.StreamExpenseResponse, error) {
		return sendCurrentExpense(ctx, s.dbClient, req.Msg.GetId())
	}, srv, &streamExpenseAlive); err != nil {
		if eris.Is(err, service.ErrResourceNoLongerFound) {
			return connect.NewError(
				connect.CodeDataLoss,
				eris.New("the expense does no longer exist"))
		} else if eris.As(err, &util.ResourceNotFoundError{}) {
			return connect.NewError(
				connect.CodeNotFound,
				eris.New("the expense does not exist"))
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

func sendCurrentExpense(ctx context.Context, dbClient bun.IDB, expenseId string) (*expensesvcv1.StreamExpenseResponse, error) {
	expenseModel, err := util.CheckResourceExists[*model.Expense](ctx, dbClient, expenseId)
	if err != nil {
		return nil, err
	}
	return &expensesvcv1.StreamExpenseResponse{
		Update: &expensesvcv1.StreamExpenseResponse_Expense{
			Expense: expenseModel.IntoProtoExpense(),
		},
	}, nil
}
