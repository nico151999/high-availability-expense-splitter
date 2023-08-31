package expense

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	expenseprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expense/v1"
	expensesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
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
		if eris.Is(err, errMarshalExpenseCreated) || eris.Is(err, errPublishExpenseCreated) {
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
		} else if eris.Is(err, errNoGroupWithId) {
			return nil, connect.NewError(connect.CodeNotFound, eris.New("the group ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensesvcv1.CreateExpenseResponse{
		ExpenseId: expenseId,
	}), nil
}

func createExpense(ctx context.Context, nc *nats.Conn, db bun.IDB, req *expensesvcv1.CreateExpenseRequest) (string, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	expenseId := util.GenerateIdWithPrefix("expense")
	requestorEmail := "ab@c.de" // TODO: take user email from context

	if err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if err := tx.NewSelect().Model(&groupv1.Group{
			Id: req.GetGroupId(),
		}).WherePK().Limit(1).Scan(ctx); err != nil {
			if eris.Is(err, sql.ErrNoRows) {
				log.Debug("group not found", logging.Error(err))
				return errNoGroupWithId
			}
			log.Error("failed getting group", logging.Error(err))
			return errSelectGroup
		}

		var name *string
		if req != nil {
			name = req.Name
		}
		if _, err := tx.NewInsert().Model(
			model.NewExpense(&expensev1.Expense{
				Id:              expenseId,
				GroupId:         req.GetGroupId(),
				Name:            name,
				By:              req.GetBy(),
				Timestamp:       req.GetTimestamp(),
				CurrencyAcronym: req.GetCurrencyAcronym(),
			}),
		).Exec(ctx); err != nil {
			log.Error("failed inserting expense", logging.Error(err))
			return errInsertExpense
		}

		marshalled, err := proto.Marshal(&expenseprocv1.ExpenseCreated{
			ExpenseId:      expenseId,
			GroupId:        req.GetGroupId(),
			Name:           req.GetName(),
			RequestorEmail: requestorEmail,
		})
		if err != nil {
			log.Error("failed marshalling expense created event", logging.Error(err))
			return errMarshalExpenseCreated
		}
		if err := nc.Publish(environment.GetExpenseCreatedSubject(req.GetGroupId(), expenseId), marshalled); err != nil {
			log.Error("failed publishing expense created event", logging.Error(err))
			return errPublishExpenseCreated
		}
		return nil
	}); err != nil {
		return "", err
	}
	return expenseId, nil
}
