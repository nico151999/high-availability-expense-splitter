package expensestake

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensestake/v1"
	expensestakesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1"
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

var streamExpenseStakeAlive = expensestakesvcv1.StreamExpenseStakeResponse{
	Update: &expensestakesvcv1.StreamExpenseStakeResponse_StillAlive{},
}

func (s *expensestakeServer) StreamExpenseStake(ctx context.Context, req *connect.Request[expensestakesvcv1.StreamExpenseStakeRequest], srv *connect.ServerStream[expensestakesvcv1.StreamExpenseStakeResponse]) error {
	ctx, cancel := context.WithTimeout(
		logging.IntoContext(
			ctx,
			logging.FromContext(ctx).With(
				logging.String(
					"expensestakeId",
					req.Msg.GetId()))),
		time.Hour)
	defer cancel()

	streamSubject := fmt.Sprintf("%s.*", environment.GetExpenseStakeSubject("*", "*", req.Msg.GetId()))
	if err := service.StreamResource(ctx, s.natsClient.Conn, streamSubject, func(ctx context.Context) (*expensestakesvcv1.StreamExpenseStakeResponse, error) {
		return sendCurrentExpenseStake(ctx, s.dbClient, req.Msg.GetId())
	}, srv, &streamExpenseStakeAlive); err != nil {
		if eris.Is(err, service.ErrResourceNoLongerFound) {
			return connect.NewError(
				connect.CodeDataLoss,
				eris.New("the expense stake does no longer exist"))
		} else if eris.As(err, &util.ResourceNotFoundError{}) {
			return connect.NewError(
				connect.CodeNotFound,
				eris.New("the expense stake does not exist"))
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

func sendCurrentExpenseStake(ctx context.Context, dbClient bun.IDB, expensestakeId string) (*expensestakesvcv1.StreamExpenseStakeResponse, error) {
	expensestake, err := util.CheckResourceExists[*expensestakev1.ExpenseStake](ctx, dbClient, expensestakeId)
	if err != nil {
		return nil, err
	}
	return &expensestakesvcv1.StreamExpenseStakeResponse{
		Update: &expensestakesvcv1.StreamExpenseStakeResponse_ExpenseStake{
			ExpenseStake: expensestake,
		},
	}, nil
}
