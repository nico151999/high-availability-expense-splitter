package category

import (
	"context"
	"time"

	"connectrpc.com/connect"
	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/category/v1"
	categorysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *categoryServer) ListCategoryIdsInGroup(ctx context.Context, req *connect.Request[categorysvcv1.ListCategoryIdsInGroupRequest]) (*connect.Response[categorysvcv1.ListCategoryIdsInGroupResponse], error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	categoryIds, err := listCategoryIds(ctx, s.dbClient, req.Msg.GetGroupId())
	if err != nil {
		if eris.Is(err, errSelectCategoryIds) {
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
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&categorysvcv1.ListCategoryIdsInGroupResponse{
		CategoryIds: categoryIds,
	}), nil
}

func listCategoryIds(ctx context.Context, dbClient bun.IDB, groupId string) ([]string, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	var categoryIds []string
	if err := dbClient.NewSelect().Model((*categoryv1.Category)(nil)).Where("group_id = ?", groupId).Column("id").Scan(ctx, &categoryIds); err != nil {
		log.Error("failed getting category IDs", logging.Error(err))
		// TODO: determine reason why category ID couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectCategoryIds
	}

	return categoryIds, nil
}
