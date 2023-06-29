package interceptor

import (
	"context"
	"fmt"
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

func UnaryValidateInterceptorFunc() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			log := logging.FromContext(ctx).NewNamed("unaryValidateInterceptorFunc")
			if err := validateMessageWithConnectError(log, req.Any(), connect.CodeInvalidArgument, "user request"); err != nil {
				return nil, err
			}
			res, err := next(ctx, req)
			if err == nil {
				if err := validateMessageWithConnectError(log, res.Any(), connect.CodeInternal, "server response"); err != nil {
					return nil, err
				}
			}
			return res, err
		})
	}
}

func validateMessageWithConnectError(log logging.Logger, msg interface{}, errorCode connect.Code, msgKind string) *connect.Error {
	violations, err := validateMessage(log, msg)
	if err != nil {
		return connect.NewError(connect.CodeInternal, eris.Errorf("failed validating %s message", msgKind))
	}
	if len(violations) == 0 {
		return nil
	}
	detail, err := connect.NewErrorDetail(&errdetails.BadRequest{
		FieldViolations: violations,
	})
	if err != nil {
		log.Error("failed to create field violation error details", logging.Error(err))
		return connect.NewError(connect.CodeInternal, eris.Errorf("failed providing error details about invalid %s message", msgKind))
	}
	conErr := connect.NewError(
		errorCode,
		eris.Errorf("the %s message is invalid", msgKind),
	)
	conErr.AddDetail(detail)
	return conErr
}

// validateMessage expects a struct to be parsed and recursively validates all fields that are vaildatable
func validateMessage(log logging.Logger, msg interface{}) ([]*errdetails.BadRequest_FieldViolation, error) {
	switch v := msg.(type) {
	case multiValidatableMessage:
		if err := v.ValidateAll(); err != nil {
			if multiErr, ok := err.(multiValidationError); ok {
				fieldViolations := []*errdetails.BadRequest_FieldViolation{}
				for _, err := range multiErr.AllErrors() {
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
	msgType := reflect.TypeOf(msg)
	for i := 0; i < msgType.Elem().NumField(); i++ {
		field := reflect.ValueOf(msg).Elem().Field(i)
		if field.Kind() == reflect.Pointer {
			details, err := validateMessage(log, field.Interface())
			if err != nil {
				return nil, err
			}
			for _, detail := range details {
				detail.Field = fmt.Sprintf("%s.%s", msgType.Field(i).Name, detail.Field)
				fieldViolations = append(fieldViolations, detail)
			}
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
