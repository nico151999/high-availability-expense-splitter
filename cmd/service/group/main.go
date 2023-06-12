package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
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
	environment.GetGlobalDomain(ctx)
	environment.GetTraceCollectorHost(ctx)
	environment.GetTraceCollectorPort(ctx)

	svc, err := group.NewGroupServer(
		fmt.Sprintf("%s:%d",
			environment.GetNatsServerHost(ctx),
			environment.GetNatsServerPort(ctx)),
		environment.GetGroupDbUser(ctx),
		environment.GetGroupDbPassword(ctx),
		fmt.Sprintf("%s:%d", environment.GetGroupDbHost(ctx), environment.GetGroupDbPort(ctx)),
		environment.GetGroupDbName(ctx))
	if err != nil {
		log.Panic(
			"failed creating new group server",
			logging.Error(err),
		)
	}
	defer svc.Close()

	corsSettings := cors.MustLoadCorsFromParam(ctx, &corsCfgFlag)
	serverAddress := fmt.Sprintf(":%d", environment.GetGroupServerPort(ctx))

	srv, err := server.ListenAndServe[groupv1connect.GroupServiceHandler](
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
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := srv.Close(ctx); err != nil {
			log.Error("failed closing server on shutdown", logging.Error(err))
		}
	}()
}
