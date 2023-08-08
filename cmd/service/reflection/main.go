package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1/groupv1connect"
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
		groupv1connect.GroupServiceName,
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
