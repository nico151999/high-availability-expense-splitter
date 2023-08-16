package server

import (
	"net/http"

	"connectrpc.com/connect"
)

type ServiceHandlerCreatorFunc[CONNECT_HANDLER any] func(svc CONNECT_HANDLER, opts ...connect.HandlerOption) (string, http.Handler)
