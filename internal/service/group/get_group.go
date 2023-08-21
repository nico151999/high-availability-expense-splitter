package group

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
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

func (s *groupServer) GetGroup(ctx context.Context, req *connect.Request[groupsvcv1.GetGroupRequest]) (*connect.Response[groupsvcv1.GetGroupResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"groupId",
				req.Msg.GetGroupId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	group, err := getGroup(ctx, s.dbClient, req.Msg.GetGroupId())
	if err != nil {
		if eris.Is(err, errSelectGroup) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBSelectErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
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

	return connect.NewResponse(&groupsvcv1.GetGroupResponse{
		Group: group,
	}), nil
}

func getGroup(ctx context.Context, dbClient bun.IDB, groupId string) (*groupv1.Group, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	group := groupv1.Group{
		Id: groupId,
	}
	if err := dbClient.NewSelect().Model(&group).WherePK().Limit(1).Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			log.Debug("group not found", logging.Error(err))
			return nil, errNoGroupWithId
		}
		log.Error("failed getting group", logging.Error(err))
		return nil, errSelectGroup
	}

	return &group, nil
}
