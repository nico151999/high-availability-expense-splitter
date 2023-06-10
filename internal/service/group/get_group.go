package group

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *groupServer) GetGroup(ctx context.Context, req *connect.Request[groupsvcv1.GetGroupRequest]) (*connect.Response[groupsvcv1.GetGroupResponse], error) {
	// TODO: tracing
	log := logging.FromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	group, err := getGroup(ctx, s.dbClient, req.Msg.GetGroupId())
	if err != nil {
		statusCode := codes.Internal
		log.Error("failed getting group from database",
			logging.String("groupId", req.Msg.GetGroupId()),
			logging.Error(err))
		st, err := status.New(statusCode, "requesting group failed").WithDetails(&errdetails.ErrorInfo{
			Reason: environment.GetDBSelectErrorReason(ctx),
			Domain: environment.GetGlobalDomain(ctx),
		})
		if err != nil {
			log.Panic("unexpected error attaching metadata", logging.Error(err))
		}
		return nil, st.Err()
	}

	return connect.NewResponse(&groupsvcv1.GetGroupResponse{
		Group: group,
	}), nil
}

func getGroup(ctx context.Context, dbClient bun.IDB, groupId string) (*groupv1.GroupProperties, error) {
	var group model.Group
	if err := dbClient.NewSelect().Model(&group).Where("group_id = ?", groupId).Limit(1).Scan(ctx); err != nil {
		// TODO: determine reason why group ID couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, eris.Wrapf(err, "failed getting group with ID %s", groupId)
	}

	return &group.GroupProperties, nil
}
