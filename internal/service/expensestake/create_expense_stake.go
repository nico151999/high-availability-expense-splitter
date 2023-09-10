package expensestake

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensestake/v1"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	expensestakeprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensestake/v1"
	expensestakesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1"
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

func (s *expensestakeServer) CreateExpenseStake(ctx context.Context, req *connect.Request[expensestakesvcv1.CreateExpenseStakeRequest]) (*connect.Response[expensestakesvcv1.CreateExpenseStakeResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expenseId",
				req.Msg.GetExpenseId()),
			logging.String(
				"forId",
				req.Msg.GetForId()),
		),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expensestakeId, err := createExpenseStake(ctx, s.natsClient, s.dbClient, req.Msg)
	if err != nil {
		if eris.Is(err, errMarshalExpenseStakeCreated) || eris.Is(err, errPublishExpenseStakeCreated) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed finalizing expense stake creation",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetMessagePublicationErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errInsertExpenseStake) {
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

	return connect.NewResponse(&expensestakesvcv1.CreateExpenseStakeResponse{
		Id: expensestakeId,
	}), nil
}

func createExpenseStake(ctx context.Context, nc *nats.Conn, db bun.IDB, req *expensestakesvcv1.CreateExpenseStakeRequest) (string, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	expensestakeId := util.GenerateIdWithPrefix("expensestake")
	requestorEmail := "ab@c.de" // TODO: take user email from context

	if err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		expense, err := util.CheckResourceExists[*model.Expense](ctx, tx, req.GetExpenseId())
		if err != nil {
			return err
		}
		if _, err := util.CheckResourceExists[*personv1.Person](ctx, tx, req.GetForId()); err != nil {
			return err
		}

		var fractionalValue *int32
		if req != nil {
			fractionalValue = req.FractionalValue
		}
		if _, err := tx.NewInsert().Model(&expensestakev1.ExpenseStake{
			Id:              expensestakeId,
			ExpenseId:       req.GetExpenseId(),
			ForId:           req.GetForId(),
			MainValue:       req.GetMainValue(),
			FractionalValue: fractionalValue,
		}).Exec(ctx); err != nil {
			log.Error("failed inserting expense stake", logging.Error(err))
			return errInsertExpenseStake
		}

		marshalled, err := proto.Marshal(&expensestakeprocv1.ExpenseStakeCreated{
			Id:              expensestakeId,
			ExpenseId:       req.GetExpenseId(),
			ForId:           req.GetForId(),
			MainValue:       req.GetMainValue(),
			FractionalValue: fractionalValue,
			RequestorEmail:  requestorEmail,
		})
		if err != nil {
			log.Error("failed marshalling expense stake created event", logging.Error(err))
			return errMarshalExpenseStakeCreated
		}
		if err := nc.Publish(environment.GetExpenseStakeCreatedSubject(expense.GetGroupId(), req.GetExpenseId(), expensestakeId), marshalled); err != nil {
			log.Error("failed publishing expense stake created event", logging.Error(err))
			return errPublishExpenseStakeCreated
		}
		return nil
	}); err != nil {
		return "", err
	}
	return expensestakeId, nil
}
