package category

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/category/v1"
	categoryprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/category/v1"
	categorysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *categoryServer) UpdateCategory(ctx context.Context, req *connect.Request[categorysvcv1.UpdateCategoryRequest]) (*connect.Response[categorysvcv1.UpdateCategoryResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"categoryId",
				req.Msg.GetCategoryId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	category, err := updateCategory(ctx, s.natsClient, s.dbClient, req.Msg.GetCategoryId(), req.Msg.GetUpdateFields())
	if err != nil {
		if eris.Is(err, errUpdateCategory) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "updating category failed",
						Domain: environment.GetDBUpdateErrorReason(ctx),
					},
				})
		} else if eris.Is(err, errNoCategoryWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the category ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&categorysvcv1.UpdateCategoryResponse{
		Category: category,
	}), nil
}

func updateCategory(ctx context.Context, nc *nats.Conn, dbClient bun.IDB, categoryId string, params []*categorysvcv1.UpdateCategoryRequest_UpdateField) (*categoryv1.Category, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	category := categoryv1.Category{
		Id: categoryId,
	}
	query := dbClient.NewUpdate()
	for _, param := range params {
		switch param.GetUpdateOption().(type) {
		case *categorysvcv1.UpdateCategoryRequest_UpdateField_Name:
			category.Name = param.GetName()
			query.Column("name")
		}
	}
	if err := query.Model(&category).WherePK().Returning("group_id").Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			log.Info("category not found", logging.Error(err))
			return nil, errNoCategoryWithId
		}
		log.Error("failed updating category", logging.Error(err))
		return nil, errUpdateCategory
	}

	marshalled, err := proto.Marshal(&categoryprocv1.CategoryUpdated{
		CategoryId: categoryId,
	})
	if err != nil {
		log.Error("failed marshalling category updated event", logging.Error(err))
		return nil, errMarshalCategoryUpdated
	}
	if err := nc.Publish(environment.GetCategoryUpdatedSubject(category.GroupId, categoryId), marshalled); err != nil {
		log.Error("failed publishing category updated event", logging.Error(err))
		return nil, errPublishCategoryUpdated
	}

	return &category, nil
}
