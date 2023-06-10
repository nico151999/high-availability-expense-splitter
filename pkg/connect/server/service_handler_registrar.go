package server

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type ServiceHandlerRegistrarFunc func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
