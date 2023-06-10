package server

import (
	"context"
	"reflect"

	"github.com/bufbuild/connect-go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
)

var ErrInvalidRequestMessage = eris.New("the passed request message is invalid")

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
			return next(ctx, req)
		})
	}
}

// validateMessage expects a struct to be parsed and recursively validates all fields that are vaildatable
func validateMessage(log logging.Logger, req interface{}) error {
	switch v := interface{}(req).(type) {
	case interface{ ValidateAll() error }:
		if err := v.ValidateAll(); err != nil {
			log.Info("invalid request message was sent by client", logging.Error(err))
			return eris.Wrap(err, "the request message is invalid")
		}
	case interface{ Validate() error }:
		if err := v.Validate(); err != nil {
			log.Info("request message with at least one invalid property was sent by client", logging.Error(err))
			return eris.Wrap(err, "the request message has at least one invalid property")
		}
	}
	for i := 0; i < reflect.TypeOf(req).Elem().NumField(); i++ {
		field := reflect.ValueOf(req).Elem().Field(i)
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
			log = log.Named(req.Spec().Procedure)
			ctx = logging.IntoContext(ctx, log)
			return next(ctx, req)
		})
	}
}

// TODO: streaming interceptors
