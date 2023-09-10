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
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
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
				req.Msg.GetId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	category, err := updateCategory(ctx, s.natsClient, s.dbClient, req.Msg.GetId(), req.Msg.GetUpdateFields())
	if err != nil {
		if eris.Is(err, errUpdateCategory) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBUpdateErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errNoCategoryWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the category ID does not exist"))
		} else if resErr := new(util.ResourceNotFoundError); eris.As(err, &resErr) {
			return nil, connect.NewError(connect.CodeNotFound, eris.Errorf("the %s with ID %s does not exist", resErr.ResourceName, resErr.ResourceId))
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

	if err := dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		query := tx.NewUpdate()
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
				return errNoCategoryWithId
			}
			log.Error("failed updating category", logging.Error(err))
			return errUpdateCategory
		}

		marshalled, err := proto.Marshal(&categoryprocv1.CategoryUpdated{
			Id:      categoryId,
			GroupId: category.GroupId,
		})
		if err != nil {
			log.Error("failed marshalling category updated event", logging.Error(err))
			return errMarshalCategoryUpdated
		}
		if err := nc.Publish(environment.GetCategoryUpdatedSubject(category.GroupId, categoryId), marshalled); err != nil {
			log.Error("failed publishing category updated event", logging.Error(err))
			return errPublishCategoryUpdated
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &category, nil
}
