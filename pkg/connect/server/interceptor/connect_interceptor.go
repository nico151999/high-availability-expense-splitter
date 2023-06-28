package interceptor

import (
	"context"
	"reflect"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
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

var ErrInvalidRequestMessage = eris.New("the passed request message is invalid")
var ErrInvalidResponseMessage = eris.New("the server produced an invalid response message")

func UnaryValidateInterceptorFunc() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			log := logging.FromContext(ctx).NewNamed("unaryValidateInterceptorFunc")
			errDetails, err := validateMessageWithErrorDetails(log, req)
			if err != nil {
				return nil, connect.NewError(
					connect.CodeInternal,
					eris.New("failed creating validating request"),
				)
				// TODO: error details
			}
			if errDetails != nil {
				conErr := connect.NewError(
					connect.CodeInvalidArgument,
					ErrInvalidRequestMessage,
				)
				conErr.AddDetail(errDetails)
				return nil, conErr
			}
			res, err := next(ctx, req)
			if err == nil {
				errDetails, err := validateMessageWithErrorDetails(log, res)
				if err != nil {
					return nil, connect.NewError(
						connect.CodeInternal,
						eris.New("failed creating proper response"),
					)
					// TODO: error details
				}
				if errDetails != nil {
					conErr := connect.NewError(
						connect.CodeInternal,
						ErrInvalidResponseMessage,
					)
					conErr.AddDetail(errDetails)
					return nil, conErr
				}
			}
			return res, err
		})
	}
}

func validateMessageWithErrorDetails(log logging.Logger, msg interface{}) (*connect.ErrorDetail, error) {
	violations, err := validateMessage(log, msg)
	if err != nil {
		return nil, err
	}
	if len(violations) == 0 {
		return nil, nil
	}
	detail, err := connect.NewErrorDetail(&errdetails.BadRequest{
		FieldViolations: violations,
	})
	if err != nil {
		msg := "failed to create field violation error details"
		log.Error(msg, logging.Error(err))
		return nil, eris.Wrap(err, msg)
	}
	return detail, nil
}

// validateMessage expects a struct to be parsed and recursively validates all fields that are vaildatable
func validateMessage(log logging.Logger, msg interface{}) ([]*errdetails.BadRequest_FieldViolation, error) {
	switch v := msg.(type) {
	case multiValidatableMessage:
		if err := v.ValidateAll(); err != nil {
			if err, ok := err.(multiValidationError); ok {
				fieldViolations := []*errdetails.BadRequest_FieldViolation{}
				for _, err := range err.AllErrors() {
					if valErr, ok := err.(validationError); ok {
						fieldViolations = append(fieldViolations, &errdetails.BadRequest_FieldViolation{
							Field:       valErr.Field(),
							Description: valErr.Reason(),
						})
					} else {
						msg := "unexpectedly got non-validationError type"
						log.Error(msg, logging.Error(err))
						return nil, eris.Wrap(err, msg)
					}
				}
				return fieldViolations, nil
			} else {
				msg := "unexpectedly got non-multiValidationError type"
				log.Error(msg, logging.Error(err))
				return nil, eris.Wrap(err, msg)
			}
		}
	case validatableMessage:
		if err := v.Validate(); err != nil {
			if valErr, ok := err.(validationError); ok {
				return []*errdetails.BadRequest_FieldViolation{
					{
						Field:       valErr.Field(),
						Description: valErr.Reason(),
					},
				}, nil
			} else {
				msg := "unexpectedly got non-validationError type"
				log.Error(msg, logging.Error(err))
				return nil, eris.Wrap(err, msg)
			}
		}
	}

	fieldViolations := []*errdetails.BadRequest_FieldViolation{}
	for i := 0; i < reflect.TypeOf(msg).Elem().NumField(); i++ {
		field := reflect.ValueOf(msg).Elem().Field(i)
		if field.Kind() == reflect.Pointer {
			details, err := validateMessage(log, field.Interface())
			if err != nil {
				return nil, err
			}
			fieldViolations = append(fieldViolations, details...)
		}
	}
	return fieldViolations, nil
}

// UnaryLogInterceptorFunc adds a named logger to the context of the request
func UnaryLogInterceptorFunc(ctx context.Context) connect.UnaryInterceptorFunc {
	log := logging.FromContext(ctx)
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			log = log.NewNamed(
				strings.ReplaceAll(req.Spec().Procedure, ".", "-"),
			)
			ctx = logging.IntoContext(ctx, log)
			return next(ctx, req)
		})
	}
}

// TODO: streaming interceptors
