package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1/personv1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/person"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

const serviceName = "personService"

func main() {
	log := logging.GetLogger().Named(serviceName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetPersonServerPort(ctx)
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
	environment.GetPeopleSubject("foo")
	environment.GetPersonSubject("foo", "bar")
	environment.GetPersonCreatedSubject("foo", "bar")
	environment.GetPersonDeletedSubject("foo", "bar")
	environment.GetPersonUpdatedSubject("foo", "bar")

	svc, err := person.NewPersonServer(
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
			"failed creating new person server",
			logging.Error(err),
		)
	}
	defer svc.Close()

	serverAddress := fmt.Sprintf(":%d", environment.GetPersonServerPort(ctx))

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err = server.ListenAndServe[personv1connect.PersonServiceHandler](
		ctx,
		serverAddress,
		svc,
		personv1.RegisterPersonServiceHandler,
		personv1connect.NewPersonServiceHandler,
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
