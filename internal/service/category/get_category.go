package category

import (
	"context"
	"database/sql"
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

func (s *categoryServer) GetCategory(ctx context.Context, req *connect.Request[categorysvcv1.GetCategoryRequest]) (*connect.Response[categorysvcv1.GetCategoryResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"categoryId",
				req.Msg.GetCategoryId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	category, err := getCategory(ctx, s.dbClient, req.Msg.GetCategoryId())
	if err != nil {
		if eris.Is(err, errSelectCategory) {
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
		} else if eris.Is(err, errNoCategoryWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the category ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&categorysvcv1.GetCategoryResponse{
		Category: category,
	}), nil
}

func getCategory(ctx context.Context, dbClient bun.IDB, categoryId string) (*categoryv1.Category, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	category := categoryv1.Category{
		Id: categoryId,
	}
	if err := dbClient.NewSelect().Model(&category).WherePK().Limit(1).Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			log.Debug("category not found", logging.Error(err))
			return nil, errNoCategoryWithId
		}
		log.Error("failed getting category", logging.Error(err))
		return nil, errSelectCategory
	}

	return &category, nil
}
