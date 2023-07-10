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

var errSelectGroup = eris.New("failed selecting group")
var errNoGroupWithId = eris.New("there is no group with that ID")

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
						Reason: "requesting group from database failed",
						Domain: environment.GetDBSelectErrorReason(ctx),
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
	var group groupv1.Group
	if err := dbClient.NewSelect().Model(&group).Where("id = ?", groupId).Limit(1).Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return nil, errNoGroupWithId
		}
		log.Error("failed getting group", logging.Error(err))
		// TODO: determine reason why group ID couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectGroup
	}

	return &group, nil
}
