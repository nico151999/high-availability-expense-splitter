package group

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
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

var streamGroupIdsAlive = groupsvcv1.StreamGroupIdsResponse{
	Update: &groupsvcv1.StreamGroupIdsResponse_StillAlive{},
}

func (s *groupServer) StreamGroupIds(ctx context.Context, req *connect.Request[groupsvcv1.StreamGroupIdsRequest], srv *connect.ServerStream[groupsvcv1.StreamGroupIdsResponse]) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	if err := service.StreamResource(ctx, s.natsClient, fmt.Sprintf("%s.*", environment.GetGroupSubject("*")), func(ctx context.Context) (*groupsvcv1.StreamGroupIdsResponse, error) {
		return sendCurrentGroupIds(ctx, s.dbClient)
	}, srv, &streamGroupIdsAlive); err != nil {
		if eris.Is(err, errSelectGroupIds) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "requesting group IDs from database failed",
						Domain: environment.GetDBSelectErrorReason(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSubscribeResource) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed subscribing to updates",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "subscribing to group ID updates failed",
						Domain: environment.GetMessageSubscriptionErrorReason(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendCurrentResourceMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed returning current resource",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "returning current group IDs failed",
						Domain: environment.GetSendCurrentResourceErrorReason(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendStreamAliveMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed sending alive message to client",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "the periodic alive check failed",
						Domain: environment.GetSendStreamAliveErrorReason(ctx),
					},
				})
		} else {
			return connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return nil
}

func sendCurrentGroupIds(ctx context.Context, dbClient bun.IDB) (*groupsvcv1.StreamGroupIdsResponse, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	var groupIds []string
	if err := dbClient.NewSelect().Model((*groupv1.Group)(nil)).Column("id").Scan(ctx, &groupIds); err != nil {
		log.Error("failed getting group IDs", logging.Error(err))
		// TODO: determine reason why group IDs couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectGroupIds
	}
	return &groupsvcv1.StreamGroupIdsResponse{
		Update: &groupsvcv1.StreamGroupIdsResponse_GroupIds_{
			GroupIds: &groupsvcv1.StreamGroupIdsResponse_GroupIds{
				GroupIds: groupIds,
			},
		},
	}, nil
}
