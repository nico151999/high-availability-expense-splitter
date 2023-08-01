package group

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
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

func (s *groupServer) StreamGroupIds(ctx context.Context, req *connect.Request[groupsvcv1.StreamGroupIdsRequest], srv *connect.ServerStream[groupsvcv1.StreamGroupIdsResponse]) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	if err := service.StreamResource(ctx, s.natsClient, environment.GetGroupSubject(), func(ctx context.Context, srv *connect.ServerStream[groupsvcv1.StreamGroupIdsResponse]) error {
		return sendCurrentGroupIds(ctx, s.dbClient, srv)
	}, srv, &groupsvcv1.StreamGroupIdsResponse{
		Update: &groupsvcv1.StreamGroupIdsResponse_StillAlive{},
	}); err != nil {
		// TODO: catch more error cases
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
		} else {
			return connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return nil
}

func sendCurrentGroupIds(ctx context.Context, dbClient bun.IDB, srv *connect.ServerStream[groupsvcv1.StreamGroupIdsResponse]) error {
	log := otel.NewOtelLoggerFromContext(ctx)

	var groupIds []string
	if err := dbClient.NewSelect().Model((*groupv1.Group)(nil)).Column("id").Scan(ctx, &groupIds); err != nil {
		log.Error("failed getting group IDs", logging.Error(err))
		// TODO: determine reason why group ID couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return errSelectGroupIds
	}
	if err := srv.Send(&groupsvcv1.StreamGroupIdsResponse{
		Update: &groupsvcv1.StreamGroupIdsResponse_GroupIds_{
			GroupIds: &groupsvcv1.StreamGroupIdsResponse_GroupIds{
				GroupIds: groupIds,
			},
		},
	}); err != nil {
		log.Error("failed sending current group IDs to client", logging.Error(err))
		return errSendStreamMessage
	}
	return nil
}
