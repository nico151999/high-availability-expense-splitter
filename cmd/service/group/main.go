package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1/groupv1connect"
	"github.com/nico151999/high-availability-expense-splitter/internal/config/cors"
	"github.com/nico151999/high-availability-expense-splitter/internal/service/group"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/param"
)

const serviceName = "groupService"

var corsCfgFlag param.StringParam

func init() {
	flag.Var(&corsCfgFlag, "groupSvcCorsCfg", "the filepath of the cors config yaml")
	flag.Parse()
}

func main() {
	log := logging.GetLogger().Named(serviceName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetGroupServerPort(ctx)
	environment.GetNatsServerHost(ctx)
	environment.GetNatsServerPort(ctx)
	environment.GetDbUser(ctx)
	environment.GetDbPassword(ctx)
	environment.GetDbHost(ctx)
	environment.GetDbPort(ctx)
	environment.GetGlobalDomain(ctx)
	environment.GetTraceCollectorHost(ctx)
	environment.GetTraceCollectorPort(ctx)
	environment.GetTaskPublicationErrorReason(ctx)
	environment.GetDBSelectErrorReason(ctx)

	svc, err := group.NewGroupServer(
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
			"failed creating new group server",
			logging.Error(err),
		)
	}
	defer svc.Close()

	corsSettings := cors.MustLoadCorsFromParam(ctx, &corsCfgFlag)
	serverAddress := fmt.Sprintf(":%d", environment.GetGroupServerPort(ctx))

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err = server.ListenAndServe[groupv1connect.GroupServiceHandler](
		ctx,
		serverAddress,
		svc,
		groupv1.RegisterGroupServiceHandler,
		groupv1connect.NewGroupServiceHandler,
		serviceName,
		fmt.Sprintf("%s:%d",
			environment.GetTraceCollectorHost(ctx),
			environment.GetTraceCollectorPort(ctx)),
		corsSettings.UrlPatterns,
		corsSettings.AllowedHeaders,
		corsSettings.AllowedMethods)
	if err != nil {
		log.Panic(
			"failed running server",
			logging.Error(err))
	}
}
