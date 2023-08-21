package category

import (
	"context"
	"database/sql"
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

var streamCategoryAlive = categorysvcv1.StreamCategoryResponse{
	Update: &categorysvcv1.StreamCategoryResponse_StillAlive{},
}

func (s *categoryServer) StreamCategory(ctx context.Context, req *connect.Request[categorysvcv1.StreamCategoryRequest], srv *connect.ServerStream[categorysvcv1.StreamCategoryResponse]) error {
	ctx, cancel := context.WithTimeout(
		logging.IntoContext(
			ctx,
			logging.FromContext(ctx).With(
				logging.String(
					"categoryId",
					req.Msg.GetCategoryId()))),
		time.Hour)
	defer cancel()

	streamSubject := fmt.Sprintf("%s.*", environment.GetCategorySubject("*", req.Msg.GetCategoryId()))
	if err := service.StreamResource(ctx, s.natsClient, streamSubject, func(ctx context.Context) (*categorysvcv1.StreamCategoryResponse, error) {
		return sendCurrentCategory(ctx, s.dbClient, req.Msg.GetCategoryId())
	}, srv, &streamCategoryAlive); err != nil {
		if eris.Is(err, service.ErrResourceNoLongerFound) {
			return connect.NewError(
				connect.CodeDataLoss,
				eris.New("the category does no longer exist"))
		} else if eris.Is(err, service.ErrResourceNotFound) {
			return connect.NewError(
				connect.CodeNotFound,
				eris.New("the category does not exist"))
		} else if eris.Is(err, errSelectCategory) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "requesting current category from database failed",
						Domain: environment.GetDBSelectErrorReason(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSubscribeResource) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed subscribing to updates",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "subscribing to category updates failed",
						Domain: environment.GetMessageSubscriptionErrorReason(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendCurrentResourceMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed returning current resource",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "returning current category failed",
						Domain: environment.GetSendCurrentResourceErrorReason(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendStreamAliveMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed sending alive message to client",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "the periodic alive check failed",
						Domain: environment.GetSendStreamAliveErrorReason(ctx),
					},
				})
		} else {
			return connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return nil
}

func sendCurrentCategory(ctx context.Context, dbClient bun.IDB, categoryId string) (*categorysvcv1.StreamCategoryResponse, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	var category categoryv1.Category
	if err := dbClient.NewSelect().Model(&category).Where("id = ?", categoryId).Limit(1).Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			log.Debug("category not found", logging.Error(err))
			return nil, service.ErrResourceNotFound
		}
		log.Error("failed getting category", logging.Error(err))
		return nil, errSelectCategory
	}
	return &categorysvcv1.StreamCategoryResponse{
		Update: &categorysvcv1.StreamCategoryResponse_Category{
			Category: &category,
		},
	}, nil
}
