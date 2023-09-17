package group

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
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

var streamGroupAlive = groupsvcv1.StreamGroupResponse{
	Update: &groupsvcv1.StreamGroupResponse_StillAlive{},
}

func (s *groupServer) StreamGroup(ctx context.Context, req *connect.Request[groupsvcv1.StreamGroupRequest], srv *connect.ServerStream[groupsvcv1.StreamGroupResponse]) error {
	ctx, cancel := context.WithTimeout(
		logging.IntoContext(
			ctx,
			logging.FromContext(ctx).With(
				logging.String(
					"groupId",
					req.Msg.GetId()))),
		time.Hour)
	defer cancel()

	if err := service.StreamResource(ctx, s.natsClient, fmt.Sprintf("%s.*", environment.GetGroupSubject(req.Msg.GetId())), func(ctx context.Context) (*groupsvcv1.StreamGroupResponse, error) {
		return sendCurrentGroup(ctx, s.dbClient, req.Msg.GetId())
	}, srv, &streamGroupAlive); err != nil {
		if eris.Is(err, service.ErrResourceNoLongerFound) {
			return connect.NewError(
				connect.CodeDataLoss,
				eris.New("the group does no longer exist"))
		} else if eris.As(err, &util.ResourceNotFoundError{}) {
			return connect.NewError(
				connect.CodeNotFound,
				eris.New("the group does not exist"))
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

func sendCurrentGroup(ctx context.Context, dbClient bun.IDB, groupId string) (*groupsvcv1.StreamGroupResponse, error) {
	group, err := util.CheckResourceExists[*groupv1.Group](ctx, dbClient, groupId)
	if err != nil {
		return nil, err
	}
	return &groupsvcv1.StreamGroupResponse{
		Update: &groupsvcv1.StreamGroupResponse_Group{
			Group: group,
		},
	}, nil
}
