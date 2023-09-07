package category

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/category/v1"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
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

func (s *categoryServer) CreateCategory(ctx context.Context, req *connect.Request[categorysvcv1.CreateCategoryRequest]) (*connect.Response[categorysvcv1.CreateCategoryResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"categoryName",
				req.Msg.GetName())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	categoryId, err := createCategory(ctx, s.natsClient, s.dbClient, req.Msg)
	if err != nil {
		if eris.Is(err, errMarshalCategoryCreated) || eris.Is(err, errPublishCategoryCreated) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed finalizing category creation",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetMessagePublicationErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errInsertCategory) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBInsertErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if resErr := new(util.ResourceNotFoundError); eris.As(err, &resErr) {
			return nil, connect.NewError(connect.CodeNotFound, eris.Errorf("the %s with ID %s does not exist", resErr.ResourceName, resErr.ResourceId))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&categorysvcv1.CreateCategoryResponse{
		Id: categoryId,
	}), nil
}

func createCategory(ctx context.Context, nc *nats.Conn, db bun.IDB, req *categorysvcv1.CreateCategoryRequest) (string, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	categoryId := util.GenerateIdWithPrefix("category")
	requestorEmail := "ab@c.de" // TODO: take user email from context

	if err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := util.CheckResourceExists[*groupv1.Group](ctx, tx, req.GetGroupId()); err != nil {
			return err
		}

		if _, err := tx.NewInsert().Model(&categoryv1.Category{
			Id:      categoryId,
			GroupId: req.GetGroupId(),
			Name:    req.GetName(),
		}).Exec(ctx); err != nil {
			log.Error("failed inserting category", logging.Error(err))
			return errInsertCategory
		}

		marshalled, err := proto.Marshal(&categoryprocv1.CategoryCreated{
			Id:             categoryId,
			GroupId:        req.GetGroupId(),
			Name:           req.GetName(),
			RequestorEmail: requestorEmail,
		})
		if err != nil {
			log.Error("failed marshalling category created event", logging.Error(err))
			return errMarshalCategoryCreated
		}
		if err := nc.Publish(environment.GetCategoryCreatedSubject(req.GetGroupId(), categoryId), marshalled); err != nil {
			log.Error("failed publishing category created event", logging.Error(err))
			return errPublishCategoryCreated
		}
		return nil
	}); err != nil {
		return "", err
	}
	return categoryId, nil
}
