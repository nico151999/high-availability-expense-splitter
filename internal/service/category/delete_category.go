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

func (s *categoryServer) DeleteCategory(ctx context.Context, req *connect.Request[categorysvcv1.DeleteCategoryRequest]) (*connect.Response[categorysvcv1.DeleteCategoryResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"categoryId",
				req.Msg.GetId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := deleteCategory(ctx, s.natsClient, s.dbClient, req.Msg.GetId()); err != nil {
		if eris.Is(err, errDeleteCategory) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBDeleteErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
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

	return connect.NewResponse(&categorysvcv1.DeleteCategoryResponse{}), nil
}

func deleteCategory(ctx context.Context, nc *nats.Conn, dbClient bun.IDB, categoryId string) error {
	log := otel.NewOtelLoggerFromContext(ctx)

	return dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		category := categoryv1.Category{
			Id: categoryId,
		}
		if err := tx.NewDelete().Model(&category).WherePK().Returning("group_id").Scan(ctx); err != nil {
			if eris.Is(err, sql.ErrNoRows) {
				log.Info("category not found", logging.Error(err))
				return errNoCategoryWithId
			}
			log.Error("failed deleting category", logging.Error(err))
			return errDeleteCategory
		}

		marshalled, err := proto.Marshal(&categoryprocv1.CategoryDeleted{
			Id: categoryId,
		})
		if err != nil {
			log.Error("failed marshalling category deleted event", logging.Error(err))
			return errMarshalCategoryDeleted
		}
		if err := nc.Publish(environment.GetCategoryDeletedSubject(category.GroupId, categoryId), marshalled); err != nil {
			log.Error("failed publishing category deleted event", logging.Error(err))
			return errPublishCategoryDeleted
		}
		return nil
	})
}
