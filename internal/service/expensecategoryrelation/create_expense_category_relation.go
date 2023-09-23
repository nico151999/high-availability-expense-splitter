package expensecategoryrelation

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/category/v1"
	expensecategoryrelationv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensecategoryrelation/v1"
	expensecategoryrelationprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensecategoryrelation/v1"
	expensecategoryrelationsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensecategoryrelation/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *expensecategoryrelationServer) CreateExpenseCategoryRelation(ctx context.Context, req *connect.Request[expensecategoryrelationsvcv1.CreateExpenseCategoryRelationRequest]) (*connect.Response[expensecategoryrelationsvcv1.CreateExpenseCategoryRelationResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expenseId",
				req.Msg.GetExpenseId()),
			logging.String(
				"categoryId",
				req.Msg.GetCategoryId()),
		),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := createExpenseCategoryRelation(ctx, s.natsClient, s.dbClient, req.Msg)
	if err != nil {
		if eris.Is(err, errPublishExpenseCategoryRelationCreated) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed finalizing expense category relation creation",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetMessagePublicationErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errInsertExpenseCategoryRelation) {
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
		} else if eris.Is(err, errExpenseAndCategoryInDifferentGroups) {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				eris.Errorf(
					"the expense with ID %s and category with ID %s need to be in the same group",
					req.Msg.GetExpenseId(),
					req.Msg.GetCategoryId()))
		} else if resErr := new(util.ResourceNotFoundError); eris.As(err, resErr) {
			return nil, connect.NewError(connect.CodeNotFound, eris.Errorf("the %s with ID %s does not exist", resErr.ResourceName, resErr.ResourceId))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensecategoryrelationsvcv1.CreateExpenseCategoryRelationResponse{}), nil
}

func createExpenseCategoryRelation(ctx context.Context, nc *nats.EncodedConn, db bun.IDB, req *expensecategoryrelationsvcv1.CreateExpenseCategoryRelationRequest) error {
	log := logging.FromContext(ctx)

	requestorEmail := "ab@c.de" // TODO: take user email from context

	if err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		expense, err := util.CheckResourceExists[*model.Expense](ctx, tx, req.GetExpenseId())
		if err != nil {
			return err
		}
		category, err := util.CheckResourceExists[*categoryv1.Category](ctx, tx, req.GetCategoryId())
		if err != nil {
			return err
		}

		if expense.GetGroupId() != category.GetGroupId() {
			log.Error("expense and category need to be in the same group in order to be able to create a relation")
			return errExpenseAndCategoryInDifferentGroups
		}

		if _, err := tx.NewInsert().Model(&expensecategoryrelationv1.ExpenseCategoryRelation{
			ExpenseId:  req.GetExpenseId(),
			CategoryId: req.GetCategoryId(),
		}).Exec(ctx); err != nil {
			log.Error("failed inserting expense category relation", logging.Error(err))
			return errInsertExpenseCategoryRelation
		}

		if err := nc.Publish(environment.GetExpenseCategoryRelationCreatedSubject(expense.GetGroupId(), req.GetExpenseId(), req.GetCategoryId()), &expensecategoryrelationprocv1.ExpenseCategoryRelationCreated{
			ExpenseId:      req.GetExpenseId(),
			CategoryId:     req.GetCategoryId(),
			RequestorEmail: requestorEmail,
		}); err != nil {
			log.Error("failed publishing expense category relation created event", logging.Error(err))
			return errPublishExpenseCategoryRelationCreated
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
