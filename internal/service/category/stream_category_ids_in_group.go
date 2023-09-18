package category

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/category/v1"
	categorysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1"
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

var streamCategoryIdsAlive = categorysvcv1.StreamCategoryIdsInGroupResponse{
	Update: &categorysvcv1.StreamCategoryIdsInGroupResponse_StillAlive{},
}

func (s *categoryServer) StreamCategoryIdsInGroup(ctx context.Context, req *connect.Request[categorysvcv1.StreamCategoryIdsInGroupRequest], srv *connect.ServerStream[categorysvcv1.StreamCategoryIdsInGroupResponse]) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	if err := service.StreamResource(ctx, s.natsClient, fmt.Sprintf("%s.*", environment.GetCategorySubject(req.Msg.GetGroupId(), "*")), func(ctx context.Context) (*categorysvcv1.StreamCategoryIdsInGroupResponse, error) {
		return sendCurrentCategoryIds(ctx, s.dbClient, req.Msg.GetGroupId())
	}, srv, &streamCategoryIdsAlive); err != nil {
		if eris.Is(err, errSelectCategoryIds) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBSelectErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSubscribeResource) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed subscribing to updates",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetMessageSubscriptionErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendCurrentResourceMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed returning current resource",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetSendCurrentResourceErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendStreamAliveMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed sending alive message to client",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetSendStreamAliveErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else {
			return connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return nil
}

func sendCurrentCategoryIds(ctx context.Context, dbClient bun.IDB, groupId string) (*categorysvcv1.StreamCategoryIdsInGroupResponse, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	var categoryIds []string
	if err := dbClient.NewSelect().Model((*categoryv1.Category)(nil)).Where("group_id = ?", groupId).Column("id").Order("name ASC").Scan(ctx, &categoryIds); err != nil {
		log.Error("failed getting category IDs", logging.Error(err))
		// TODO: determine reason why category IDs couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectCategoryIds
	}
	return &categorysvcv1.StreamCategoryIdsInGroupResponse{
		Update: &categorysvcv1.StreamCategoryIdsInGroupResponse_Ids{
			Ids: &categorysvcv1.StreamCategoryIdsInGroupResponse_CategoryIds{
				Ids: categoryIds,
			},
		},
	}, nil
}
