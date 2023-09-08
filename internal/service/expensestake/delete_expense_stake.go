package expensestake

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expensestake/v1"
	expensestakeprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensestake/v1"
	expensestakesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1"
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

func (s *expensestakeServer) DeleteExpenseStake(ctx context.Context, req *connect.Request[expensestakesvcv1.DeleteExpenseStakeRequest]) (*connect.Response[expensestakesvcv1.DeleteExpenseStakeResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"expensestakeId",
				req.Msg.GetId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := deleteExpenseStake(ctx, s.natsClient, s.dbClient, req.Msg.GetId()); err != nil {
		if eris.Is(err, errDeleteExpenseStake) {
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
		} else if eris.Is(err, errNoExpenseStakeWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the expense stake ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&expensestakesvcv1.DeleteExpenseStakeResponse{}), nil
}

func deleteExpenseStake(ctx context.Context, nc *nats.Conn, dbClient bun.IDB, expensestakeId string) error {
	log := otel.NewOtelLoggerFromContext(ctx)

	return dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		expensestake := expensestakev1.ExpenseStake{
			Id: expensestakeId,
		}
		if err := tx.NewDelete().Model(&expensestake).WherePK().Returning("expense_id").Scan(ctx); err != nil {
			if eris.Is(err, sql.ErrNoRows) {
				log.Info("expense stake not found", logging.Error(err))
				return errNoExpenseStakeWithId
			}
			log.Error("failed deleting expense stake", logging.Error(err))
			return errDeleteExpenseStake
		}
		expense, err := util.CheckResourceExists[*expensev1.Expense](ctx, tx, expensestake.GetExpenseId())
		if err != nil {
			return err
		}

		marshalled, err := proto.Marshal(&expensestakeprocv1.ExpenseStakeDeleted{
			Id: expensestakeId,
		})
		if err != nil {
			log.Error("failed marshalling expense stake deleted event", logging.Error(err))
			return errMarshalExpenseStakeDeleted
		}
		if err := nc.Publish(environment.GetExpenseStakeDeletedSubject(expense.GetGroupId(), expense.GetId(), expensestakeId), marshalled); err != nil {
			log.Error("failed publishing expense stake deleted event", logging.Error(err))
			return errPublishExpenseStakeDeleted
		}

		return nil
	})
}
