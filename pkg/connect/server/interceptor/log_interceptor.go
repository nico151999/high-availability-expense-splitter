package interceptor

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
)

// NewLogInterceptor creates a connect interceptor that adds a named logger to the context of each the request and logs events
func NewLogInterceptor(ctx context.Context) *logInterceptor {
	return &logInterceptor{
		log: logging.FromContext(ctx),
	}
}

var _ connect.Interceptor = (*logInterceptor)(nil)

type logInterceptor struct {
	log logging.Logger
}

func (i *logInterceptor) prepareLog(ctx context.Context, proc string) (context.Context, logging.Logger) {
	log := i.log.Named(
		strings.ReplaceAll(proc, ".", "-"),
	).WithInterceptors(
		otel.NewOtelInterceptorFunc(ctx),
	)
	return logging.IntoContext(ctx, log), log
}

func (i *logInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		ctx, log := i.prepareLog(ctx, req.Spec().Procedure)
		log.Info("received request")
		res, err := next(ctx, req)
		if err != nil {
			log.Info("request is answered with an error")
		}
		return res, err
	})
}

// WrapStreamingClient does nothing since this interceptor is a server only implementation
func (i *logInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *logInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, log := i.prepareLog(ctx, conn.Spec().Procedure)
		log.Info("starting to handle stream")
		err := next(ctx, &streamingLogHandlerConn{
			StreamingHandlerConn: conn,
			log:                  log,
		})
		if err != nil {
			log.Info("stream ended with an error", logging.Error(err))
		}
		return err
	}
}

type streamingLogHandlerConn struct {
	connect.StreamingHandlerConn

	log logging.Logger
}

func (p *streamingLogHandlerConn) Receive(msg any) error {
	p.log.Info("receiving a message")
	return p.StreamingHandlerConn.Receive(msg)
}

func (p *streamingLogHandlerConn) Send(msg any) error {
	p.log.Info("sending a message")
	return p.StreamingHandlerConn.Send(msg)
}
