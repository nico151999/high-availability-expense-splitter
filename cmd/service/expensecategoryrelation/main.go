package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	expensecategoryrelationv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensecategoryrelation/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensecategoryrelation/v1/expensecategoryrelationv1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/expensecategoryrelation"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

const serviceName = "expensecategoryrelationService"

func main() {
	log := logging.GetLogger().Named(serviceName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetExpensecategoryrelationServerPort(ctx)
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
	environment.GetExpenseCategoryRelationsSubject("foo", "bar")
	environment.GetExpenseCategoryRelationSubject("foo", "bar", "bob")
	environment.GetExpenseCategoryRelationCreatedSubject("foo", "bar", "bob")
	environment.GetExpenseCategoryRelationDeletedSubject("foo", "bar", "bob")
	environment.GetExpenseCategoryRelationUpdatedSubject("foo", "bar", "bob")

	svc, err := expensecategoryrelation.NewExpenseCategoryRelationServer(
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
			"failed creating new expensecategoryrelation server",
			logging.Error(err),
		)
	}
	defer svc.Close()

	serverAddress := fmt.Sprintf(":%d", environment.GetExpensecategoryrelationServerPort(ctx))

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err = server.ListenAndServe[expensecategoryrelationv1connect.ExpenseCategoryRelationServiceHandler](
		ctx,
		serverAddress,
		svc,
		expensecategoryrelationv1.RegisterExpenseCategoryRelationServiceHandler,
		expensecategoryrelationv1connect.NewExpenseCategoryRelationServiceHandler,
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
