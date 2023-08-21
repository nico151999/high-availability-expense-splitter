package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1/categoryv1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/category"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

const serviceName = "categoryService"

func main() {
	log := logging.GetLogger().Named(serviceName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetCategoryServerPort(ctx)
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
	environment.GetCategoriesSubject("foo")
	environment.GetCategorySubject("foo", "bar")
	environment.GetCategoryCreatedSubject("foo", "bar")
	environment.GetCategoryDeletedSubject("foo", "bar")
	environment.GetCategoryUpdatedSubject("foo", "bar")

	svc, err := category.NewCategoryServer(
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
			"failed creating new category server",
			logging.Error(err),
		)
	}
	defer svc.Close()

	serverAddress := fmt.Sprintf(":%d", environment.GetCategoryServerPort(ctx))

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err = server.ListenAndServe[categoryv1connect.CategoryServiceHandler](
		ctx,
		serverAddress,
		svc,
		categoryv1.RegisterCategoryServiceHandler,
		categoryv1connect.NewCategoryServiceHandler,
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
