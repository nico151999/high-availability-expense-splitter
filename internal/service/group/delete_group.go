package group

import (
	"context"
	"database/sql"
	"time"

	"github.com/bufbuild/connect-go"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *groupServer) DeleteGroup(ctx context.Context, req *connect.Request[groupsvcv1.DeleteGroupRequest]) (*connect.Response[groupsvcv1.DeleteGroupResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"groupId",
				req.Msg.GetGroupId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := deleteGroup(ctx, s.dbClient, req.Msg.GetGroupId()); err != nil {
		if eris.Is(err, errDeleteGroup) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "deleting group from database failed",
						Domain: environment.GetDBDeleteErrorReason(ctx),
					},
				})
		} else if eris.Is(err, errNoGroupWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the group ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&groupsvcv1.DeleteGroupResponse{}), nil
}

func deleteGroup(ctx context.Context, dbClient bun.IDB, groupId string) error {
	log := otel.NewOtelLoggerFromContext(ctx)
	var group groupv1.Group
	if _, err := dbClient.NewDelete().Model(&group).Where("id = ?", groupId).Exec(ctx); err != nil {
		log.Error("failed deleting group", logging.Error(err))
		if eris.Is(err, sql.ErrNoRows) {
			return errNoGroupWithId
		}
		return errDeleteGroup
	}

	return nil
}
