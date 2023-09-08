package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1/expensestakev1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/expensestake"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

const serviceName = "expensestakeService"

func main() {
	log := logging.GetLogger().Named(serviceName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetExpenseStakeServerPort(ctx)
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
	environment.GetExpenseStakesSubject("foo", "bar")
	environment.GetExpenseStakeSubject("foo", "bar", "bob")
	environment.GetExpenseStakeCreatedSubject("foo", "bar", "bob")
	environment.GetExpenseStakeDeletedSubject("foo", "bar", "bob")
	environment.GetExpenseStakeUpdatedSubject("foo", "bar", "bob")

	svc, err := expensestake.NewExpenseStakeServer(
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
			"failed creating new expensestake server",
			logging.Error(err),
		)
	}
	defer svc.Close()

	serverAddress := fmt.Sprintf(":%d", environment.GetExpenseStakeServerPort(ctx))

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err = server.ListenAndServe[expensestakev1connect.ExpenseStakeServiceHandler](
		ctx,
		serverAddress,
		svc,
		expensestakev1.RegisterExpenseStakeServiceHandler,
		expensestakev1connect.NewExpenseStakeServiceHandler,
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
