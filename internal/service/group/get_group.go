package group

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
)

var errSelectGroup = eris.New("failed selecting group")

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
		var conError *connect.Error
		if eris.Is(err, errSelectGroup) {
			conError = server.CreateErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"requesting group from database failed",
				environment.GetDBSelectErrorReason(ctx))
		} else {
			conError = connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
		return nil, conError
	}

	return connect.NewResponse(&groupsvcv1.GetGroupResponse{
		Group: group,
	}), nil
}

func getGroup(ctx context.Context, dbClient bun.IDB, groupId string) (*groupv1.GroupProperties, error) {
	log := otel.NewOtelLogger(ctx, logging.FromContext(ctx))
	var group model.Group
	if err := dbClient.NewSelect().Model(&group).Where("group_id = ?", groupId).Limit(1).Scan(ctx); err != nil {
		log.Error("failed getting group", logging.Error(err))
		// TODO: determine reason why group ID couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectGroup
	}

	return &group.GroupProperties, nil
}
