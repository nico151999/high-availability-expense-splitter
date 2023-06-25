package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func CreateErrorWithDetails(ctx context.Context, code connect.Code, msg, reason string) *connect.Error {
	log := otel.NewOtelLogger(ctx, logging.FromContext(ctx))

	conErr := connect.NewError(code, eris.New(msg))
	conErrDetail, err := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: reason,
		Domain: environment.GetGlobalDomain(ctx),
	})
	if err != nil {
		log.Error("unexpected error creating error metadata", logging.Error(err))
		return connect.NewError(code, eris.New("failed creating error details"))
	}
	conErr.AddDetail(conErrDetail)
	return conErr
}
