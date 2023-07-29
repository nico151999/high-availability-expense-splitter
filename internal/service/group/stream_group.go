package group

import (
	"context"
	"database/sql"

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

func (s *groupServer) StreamGroup(ctx context.Context, req *connect.Request[groupsvcv1.StreamGroupRequest], srv *connect.ServerStream[groupsvcv1.StreamGroupResponse]) error {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"groupId",
				req.Msg.GetGroupId())))

	if err := service.StreamResource(ctx, s.natsClient, environment.GetGroupSubject(), func(ctx context.Context, srv *connect.ServerStream[groupsvcv1.StreamGroupResponse]) error {
		return sendCurrentGroup(ctx, s.dbClient, req.Msg.GetGroupId(), srv)
	}, srv, &groupsvcv1.StreamGroupResponse{
		Update: &groupsvcv1.StreamGroupResponse_StillAlive{},
	}); err != nil {
		// TODO: catch more error cases
		if eris.Is(err, errSelectGroup) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "requesting current group from database failed",
						Domain: environment.GetDBSelectErrorReason(ctx),
					},
				})
		} else if eris.Is(err, errNoGroupWithId) {
			return connect.NewError(
				connect.CodeNotFound,
				eris.New("the group ID does not exist"))
		} else {
			return connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return nil
}

func sendCurrentGroup(ctx context.Context, dbClient bun.IDB, groupId string, srv *connect.ServerStream[groupsvcv1.StreamGroupResponse]) error {
	log := otel.NewOtelLoggerFromContext(ctx)

	var group groupv1.Group
	if err := dbClient.NewSelect().Model(&group).Where("id = ?", groupId).Limit(1).Scan(ctx); err != nil {
		log.Error("failed getting group", logging.Error(err))
		if eris.Is(err, sql.ErrNoRows) {
			return errNoGroupWithId
		}
		return errSelectGroup
	}
	if err := srv.Send(&groupsvcv1.StreamGroupResponse{
		Update: &groupsvcv1.StreamGroupResponse_Group{
			Group: &group,
		},
	}); err != nil {
		log.Error("failed sending current group to client", logging.Error(err))
		return errSendStreamMessage
	}
	return nil
}