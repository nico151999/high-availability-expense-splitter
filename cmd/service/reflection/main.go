package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"connectrpc.com/connect"
	grpcreflect "connectrpc.com/grpcreflect"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1/categoryv1connect"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/currency/v1/currencyv1connect"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1/expensev1connect"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensecategoryrelation/v1/expensecategoryrelationv1connect"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1/expensestakev1connect"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1/groupv1connect"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1/personv1connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"google.golang.org/grpc"
)

const serviceName = "reflectionService"

func main() {
	log := logging.GetLogger().Named(serviceName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetTraceCollectorHost(ctx)
	environment.GetTraceCollectorPort(ctx)

	serverAddress := fmt.Sprintf(":%d", environment.GetReflectionServerPort(ctx))

	svc := grpcreflect.NewStaticReflector(
		categoryv1connect.CategoryServiceName,
		currencyv1connect.CurrencyServiceName,
		expensev1connect.ExpenseServiceName,
		expensecategoryrelationv1connect.ExpenseCategoryRelationServiceName,
		expensestakev1connect.ExpenseStakeServiceName,
		groupv1connect.GroupServiceName,
		personv1connect.PersonServiceName,
	)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err := server.ListenAndServe(
		ctx,
		serverAddress,
		svc,
		func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			// we do not offer gRPC reflection via REST
			return nil
		},
		func(reflector *grpcreflect.Reflector, options ...connect.HandlerOption) (string, http.Handler) {
			mux := http.NewServeMux()
			mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector, options...))
			mux.Handle(grpcreflect.NewHandlerV1(reflector, options...))
			return "/", mux
		},
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
