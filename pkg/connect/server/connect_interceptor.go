package server

import (
	"context"
	"reflect"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
)

var ErrInvalidRequestMessage = eris.New("the passed request message is invalid")
var ErrInvalidResponseMessage = eris.New("the server produced an invalid response message")

func unaryValidateInterceptorFunc() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			log := logging.FromContext(ctx).Named("unaryValidateInterceptorFunc")
			if err := validateMessage(log, req); err != nil {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					ErrInvalidRequestMessage,
				)
			}
			res, err := next(ctx, req)
			if err == nil {
				if err := validateMessage(log, req); err != nil {
					return nil, connect.NewError(
						connect.CodeInternal,
						ErrInvalidResponseMessage,
					)
				}
			}
			return res, err
		})
	}
}

// validateMessage expects a struct to be parsed and recursively validates all fields that are vaildatable
func validateMessage(log logging.Logger, msg interface{}) error {
	switch v := interface{}(msg).(type) {
	case interface{ ValidateAll() error }:
		if err := v.ValidateAll(); err != nil {
			log.Error("invalid message was recognised", logging.Error(err))
			return eris.Wrap(err, "the message is invalid")
		}
	case interface{ Validate() error }:
		if err := v.Validate(); err != nil {
			log.Error("message with at least one invalid property was recognised", logging.Error(err))
			return eris.Wrap(err, "the message has at least one invalid property")
		}
	}
	for i := 0; i < reflect.TypeOf(msg).Elem().NumField(); i++ {
		field := reflect.ValueOf(msg).Elem().Field(i)
		if field.Kind() == reflect.Pointer {
			if err := validateMessage(log, field.Interface()); err != nil {
				return eris.Wrap(err, "a message field is invalid")
			}
		}
	}
	return nil
}

// unaryLogInterceptorFunc adds a named logger to the context of the request
func unaryLogInterceptorFunc(ctx context.Context) connect.UnaryInterceptorFunc {
	log := logging.FromContext(ctx)
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			log = log.Named(
				strings.ReplaceAll(
					strings.ReplaceAll(req.Spec().Procedure, ".", "-"),
					"/", "_",
				),
			)
			ctx = logging.IntoContext(ctx, log)
			return next(ctx, req)
		})
	}
}

// TODO: streaming interceptors
