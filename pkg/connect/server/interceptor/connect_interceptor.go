package interceptor

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type validationError interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
	Error() string
}

type multiValidationError interface {
	Error() string
	AllErrors() []error
}

type validatableMessage interface {
	Validate() error
}

type multiValidatableMessage interface {
	ValidateAll() error
}

func UnaryValidateInterceptorFunc() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			log := logging.FromContext(ctx).NewNamed("unaryValidateInterceptorFunc")
			violations, err := validateMessage(log, req.Any())
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, eris.New("failed validating request message"))
			}
			if len(violations) != 0 {
				detail, err := connect.NewErrorDetail(&errdetails.BadRequest{
					FieldViolations: violations,
				})
				if err != nil {
					log.Error("failed to create field violation error details", logging.Error(err))
					return nil, connect.NewError(connect.CodeInternal, eris.New("failed providing error details about invalid request message"))
				}
				conErr := connect.NewError(
					connect.CodeInvalidArgument,
					eris.New("the request message is invalid"),
				)
				conErr.AddDetail(detail)
				return nil, conErr
			}
			res, err := next(ctx, req)
			if err == nil {
				violations, err := validateMessage(log, res.Any())
				if err != nil {
					return nil, connect.NewError(connect.CodeInternal, eris.New("failed checking server response message"))
				}
				if len(violations) != 0 {
					for _, violation := range violations {
						log.Error(
							"invalid field in response message",
							logging.String("field", violation.GetField()),
							logging.String("description", violation.GetDescription()),
						)
					}
					return nil, connect.NewError(connect.CodeInternal, eris.New("failed creating server response message"))
				}
			}
			return res, err
		})
	}
}

// validateMessage expects a struct to be passed and recursively validates all fields that are vaildatable
func validateMessage(log logging.Logger, msg interface{}) ([]*errdetails.BadRequest_FieldViolation, error) {
	fieldViolations := []*errdetails.BadRequest_FieldViolation{}
	switch v := msg.(type) {
	case multiValidatableMessage:
		if err := v.ValidateAll(); err != nil {
			if multiErr, ok := err.(multiValidationError); ok {
				for _, err := range multiErr.AllErrors() {
					fieldViolations, err = addFieldViolation(log, fieldViolations, err)
					if err != nil {
						return nil, err
					}
				}
			} else {
				msg := "unexpectedly got non-multiValidationError type"
				log.Error(msg, logging.Error(err))
				return nil, eris.Wrap(err, msg)
			}
		}
	case validatableMessage:
		if err := v.Validate(); err != nil {
			fieldViolations, err = addFieldViolation(log, fieldViolations, err)
			if err != nil {
				return nil, err
			}
		}
	}

	msgType := reflect.TypeOf(msg)
	for i := 0; i < msgType.Elem().NumField(); i++ {
		fieldVal := reflect.ValueOf(msg).Elem().Field(i)
		if fieldVal.Kind() == reflect.Pointer && fieldVal.Elem().Kind() == reflect.Struct {
			violations, err := validateMessage(log, fieldVal.Interface())
			if err != nil {
				return nil, err
			}
			for _, violation := range violations {
				violation.Field = fmt.Sprintf("%s.%s", msgType.Field(i).Name, violation.GetField())
				fieldViolations = append(fieldViolations, violation)
			}
		}
	}
	return fieldViolations, nil
}

func addFieldViolation(log logging.Logger, fieldViolations []*errdetails.BadRequest_FieldViolation, err error) ([]*errdetails.BadRequest_FieldViolation, error) {
	if valErr, ok := err.(validationError); ok {
		return append(fieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       valErr.Field(),
			Description: valErr.Reason(),
		}), nil
	} else {
		msg := "unexpectedly got non-validationError type"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
}

// UnaryLogInterceptorFunc adds a named logger to the context of the request
func UnaryLogInterceptorFunc(ctx context.Context) connect.UnaryInterceptorFunc {
	log := logging.FromContext(ctx)
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			traceId := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
			log = log.With(logging.Trace(traceId)).Named(
				strings.ReplaceAll(req.Spec().Procedure, ".", "-"),
			)
			log.Info("received request")
			ctx = logging.IntoContext(ctx, log)
			res, err := next(ctx, req)
			if err != nil {
				log.Info("request is answered with an error")
			}
			return res, err
		})
	}
}

// TODO: streaming interceptors
