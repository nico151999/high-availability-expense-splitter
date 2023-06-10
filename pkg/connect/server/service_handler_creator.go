package server

import (
	"net/http"

	"github.com/bufbuild/connect-go"
)

type ServiceHandlerCreatorFunc[CONNECT_HANDLER any] func(svc CONNECT_HANDLER, opts ...connect.HandlerOption) (string, http.Handler)
