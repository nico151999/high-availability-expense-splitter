package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1/expensev1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/expense"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

const serviceName = "expenseService"

func main() {
	log := logging.GetLogger().Named(serviceName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetExpenseServerPort(ctx)
	environment.GetNatsServerHost(ctx)
	environment.GetNatsServerPort(ctx)
	environment.GetDbUser(ctx)
	environment.GetDbPassword(ctx)
	environment.GetDbHost(ctx)
	environment.GetDbPort(ctx)
	environment.GetGlobalDomain(ctx)
	environment.GetTraceCollectorHost(ctx)
	environment.GetTraceCollectorPort(ctx)
	environment.GetMessagePublicationErrorReason(ctx)
	environment.GetDBSelectErrorReason(ctx)
	environment.GetDBDeleteErrorReason(ctx)
	environment.GetDBInsertErrorReason(ctx)
	environment.GetDBUpdateErrorReason(ctx)
	environment.GetMessageSubscriptionErrorReason(ctx)
	environment.GetSendCurrentResourceErrorReason(ctx)
	environment.GetSendStreamAliveErrorReason(ctx)
	environment.GetExpensesSubject("foo")
	environment.GetExpenseSubject("foo", "bar")
	environment.GetExpenseCreatedSubject("foo", "bar")
	environment.GetExpenseDeletedSubject("foo", "bar")
	environment.GetExpenseUpdatedSubject("foo", "bar")

	svc, err := expense.NewExpenseServer(
		ctx,
		fmt.Sprintf("%s:%d",
			environment.GetNatsServerHost(ctx),
			environment.GetNatsServerPort(ctx)),
		environment.GetDbUser(ctx),
		environment.GetDbPassword(ctx),
		fmt.Sprintf("%s:%d", environment.GetDbHost(ctx), environment.GetDbPort(ctx)),
		environment.GetDbName(ctx))
	if err != nil {
		log.Panic(
			"failed creating new expense server",
			logging.Error(err),
		)
	}
	defer svc.Close()

	serverAddress := fmt.Sprintf(":%d", environment.GetExpenseServerPort(ctx))

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err = server.ListenAndServe[expensev1connect.ExpenseServiceHandler](
		ctx,
		serverAddress,
		svc,
		expensev1.RegisterExpenseServiceHandler,
		expensev1connect.NewExpenseServiceHandler,
		serviceName,
		fmt.Sprintf("%s:%d",
			environment.GetTraceCollectorHost(ctx),
			environment.GetTraceCollectorPort(ctx)))
	if err != nil {
		log.Panic(
			"failed running server",
			logging.Error(err))
	}
}
