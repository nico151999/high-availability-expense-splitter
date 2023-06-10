package group

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *groupServer) GetGroupIds(ctx context.Context, req *connect.Request[groupsvcv1.GetGroupIdsRequest]) (*connect.Response[groupsvcv1.GetGroupIdsResponse], error) {
	// TODO: tracing
	log := logging.FromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	groupIds, err := getGroupIds(ctx, s.dbClient)
	if err != nil {
		statusCode := codes.Internal
		if s, ok := status.FromError(eris.Cause(err)); ok {
			statusCode = s.Code()
		}
		log.Error("failed getting group IDs from Kube API with in-cluster config", logging.Error(err))
		st, err := status.New(statusCode, "requesting group IDs from k8s failed").WithDetails(&errdetails.ErrorInfo{
			Reason: environment.GetDBSelectErrorReason(ctx),
			Domain: environment.GetGlobalDomain(ctx),
		})
		if err != nil {
			log.Panic("unexpected error attaching metadata", logging.Error(err))
		}
		return nil, st.Err()
	}

	return connect.NewResponse(&groupsvcv1.GetGroupIdsResponse{
		GroupIds: groupIds,
	}), nil
}

func getGroupIds(ctx context.Context, dbClient bun.IDB) ([]string, error) {
	var groupIds []string
	if err := dbClient.NewSelect().Model((*groupv1.GroupProperties)(nil)).Column("groupId").Scan(ctx, &groupIds); err != nil {
		return nil, eris.Wrap(err, "failed getting group IDs")
	}

	return groupIds, nil
}
