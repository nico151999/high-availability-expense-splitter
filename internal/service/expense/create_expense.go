package expense

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	expenseprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	expensesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
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

func (s *expenseServer) CreateExpense(ctx context.Context, req *connect.Request[expensesvcv1.CreateExpenseRequest]) (*connect.Response[expensesvcv1.CreateExpenseResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expenseName",
				req.Msg.GetName())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expenseId, err := createExpense(ctx, s.natsClient, s.dbClient, req.Msg)
	if err != nil {
		if eris.Is(err, errPublishExpenseCreated) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed finalizing expense creation",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetMessagePublicationErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errInsertExpense) {
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
		} else if resErr := new(util.ResourceNotFoundError); eris.As(err, resErr) {
			return nil, connect.NewError(connect.CodeNotFound, eris.Errorf("the %s with ID %s does not exist", resErr.ResourceName, resErr.ResourceId))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensesvcv1.CreateExpenseResponse{
		Id: expenseId,
	}), nil
}

func createExpense(ctx context.Context, nc *nats.EncodedConn, db bun.IDB, req *expensesvcv1.CreateExpenseRequest) (string, error) {
	log := logging.FromContext(ctx)

	expenseId := util.GenerateIdWithPrefix("expense")
	requestorEmail := "ab@c.de" // TODO: take user email from context

	if err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := util.CheckResourceExists[*groupv1.Group](ctx, tx, req.GetGroupId()); err != nil {
			return err
		}
		if _, err := util.CheckResourceExists[*personv1.Person](ctx, tx, req.GetById()); err != nil {
			return err
		}
		// TODO: also check if currency exists

		var name *string
		if req != nil {
			name = req.Name
		}
		if _, err := tx.NewInsert().Model(
			model.NewExpense(&expensev1.Expense{
				Id:         expenseId,
				GroupId:    req.GetGroupId(),
				Name:       name,
				ById:       req.GetById(),
				Timestamp:  req.GetTimestamp(),
				CurrencyId: req.GetCurrencyId(),
			}),
		).Exec(ctx); err != nil {
			log.Error("failed inserting expense", logging.Error(err))
			return errInsertExpense
		}

		if err := nc.Publish(environment.GetExpenseCreatedSubject(req.GetGroupId(), expenseId), &expenseprocv1.ExpenseCreated{
			Id:             expenseId,
			GroupId:        req.GetGroupId(),
			Name:           name,
			RequestorEmail: requestorEmail,
		}); err != nil {
			log.Error("failed publishing expense created event", logging.Error(err))
			return errPublishExpenseCreated
		}
		return nil
	}); err != nil {
		return "", err
	}
	return expenseId, nil
}
