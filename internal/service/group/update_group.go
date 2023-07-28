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

func (s *groupServer) UpdateGroup(ctx context.Context, req *connect.Request[groupsvcv1.UpdateGroupRequest]) (*connect.Response[groupsvcv1.UpdateGroupResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"groupId",
				req.Msg.GetGroupId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	group, err := updateGroup(ctx, s.dbClient, req.Msg.GetGroupId(), req.Msg.GetUpdateFields())
	if err != nil {
		if eris.Is(err, errUpdateGroup) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "updating group failed",
						Domain: environment.GetDBUpdateErrorReason(ctx),
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

	return connect.NewResponse(&groupsvcv1.UpdateGroupResponse{
		Group: group,
	}), nil
}

func updateGroup(ctx context.Context, dbClient bun.IDB, groupId string, params []*groupsvcv1.UpdateGroupRequest_UpdateField) (*groupv1.Group, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	var group groupv1.Group
	query := dbClient.NewUpdate()
	for _, param := range params {
		switch param.GetUpdateOption().(type) {
		case *groupsvcv1.UpdateGroupRequest_UpdateField_Name:
			group.Name = param.GetName()
			query.Column("name")
		}
	}
	if _, err := query.Model(&group).Where("id = ?", groupId).Exec(ctx); err != nil {
		log.Error("failed updating group", logging.Error(err))
		if eris.Is(err, sql.ErrNoRows) {
			return nil, errNoGroupWithId
		}
		return nil, errUpdateGroup
	}

	return &group, nil
}
