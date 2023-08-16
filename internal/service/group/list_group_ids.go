package group

import (
	"context"
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

func (s *groupServer) ListGroupIds(ctx context.Context, req *connect.Request[groupsvcv1.ListGroupIdsRequest]) (*connect.Response[groupsvcv1.ListGroupIdsResponse], error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	groupIds, err := listGroupIds(ctx, s.dbClient)
	if err != nil {
		if eris.Is(err, errSelectGroupIds) {
			return nil, errors.NewErrorWithDetails(
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
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&groupsvcv1.ListGroupIdsResponse{
		GroupIds: groupIds,
	}), nil
}

func listGroupIds(ctx context.Context, dbClient bun.IDB) ([]string, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	var groupIds []string
	if err := dbClient.NewSelect().Model((*groupv1.Group)(nil)).Column("id").Scan(ctx, &groupIds); err != nil {
		log.Error("failed getting group IDs", logging.Error(err))
		// TODO: determine reason why group ID couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectGroupIds
	}

	return groupIds, nil
}
