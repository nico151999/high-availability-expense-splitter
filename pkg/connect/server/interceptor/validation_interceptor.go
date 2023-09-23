package interceptor

import (
	"context"
	"fmt"
	"reflect"

	"connectrpc.com/connect"
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

func validateRequest(log logging.Logger, msg interface{}) error {
	violations, err := validateMessage(log, msg)
	if err != nil {
		return connect.NewError(connect.CodeInternal, eris.New("failed validating incoming message"))
	}
	if len(violations) != 0 {
		detail, err := connect.NewErrorDetail(&errdetails.BadRequest{
			FieldViolations: violations,
		})
		if err != nil {
			log.Error("failed to create field violation error details", logging.Error(err))
			return connect.NewError(connect.CodeInternal, eris.New("failed providing error details about invalid incoming message"))
		}
		conErr := connect.NewError(
			connect.CodeInvalidArgument,
			eris.New("the incoming message is invalid"),
		)
		conErr.AddDetail(detail)
		return conErr
	}
	return nil
}

func validateResponse(log logging.Logger, msg interface{}) error {
	violations, err := validateMessage(log, msg)
	if err != nil {
		return connect.NewError(connect.CodeInternal, eris.New("failed checking server response message"))
	}
	if len(violations) != 0 {
		for _, violation := range violations {
			log.Error(
				"invalid field in response message",
				logging.String("field", violation.GetField()),
				logging.String("description", violation.GetDescription()),
			)
		}
		return connect.NewError(connect.CodeInternal, eris.New("failed creating server response message"))
	}
	return nil
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

	msgType := reflect.TypeOf(msg).Elem()
	for i := 0; i < msgType.NumField(); i++ {
		fieldVal := reflect.ValueOf(msg).Elem().Field(i)
		if fieldVal.Kind() == reflect.Pointer && fieldVal.Elem().Kind() == reflect.Struct {
			violations, err := validateMessage(log, fieldVal.Interface())
			if err != nil {
				return nil, err
			}
			for _, violation := range violations {
				log.Info(msgType.Field(i).Name)
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

// NewValidationInterceptor creates a connect interceptor that validates messages exchanged between server and client
func NewValidationInterceptor(ctx context.Context) *validationInterceptor {
	return &validationInterceptor{
		log: logging.FromContext(ctx).Named("unaryValidateInterceptorFunc"),
	}
}

var _ connect.Interceptor = (*validationInterceptor)(nil)

type validationInterceptor struct {
	log logging.Logger
}

func (i *validationInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		if err := validateRequest(i.log, req.Any()); err != nil {
			return nil, err
		}
		res, err := next(ctx, req)
		if err == nil {
			if err := validateResponse(i.log, res.Any()); err != nil {
				return nil, err
			}
		}
		return res, err
	})
}

// WrapStreamingClient does nothing since this interceptor is a server only implementation
func (i *validationInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *validationInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, &streamingValidationHandlerConn{
			StreamingHandlerConn: conn,
			log:                  i.log,
		})
	}
}

type streamingValidationHandlerConn struct {
	connect.StreamingHandlerConn

	log logging.Logger
}

func (p *streamingValidationHandlerConn) Receive(msg any) error {
	if err := p.StreamingHandlerConn.Receive(msg); err != nil {
		p.log.Error("failed receiving message", logging.Error(err))
		return connect.NewError(connect.CodeInternal, eris.New("failed receiving message"))
	}
	return validateRequest(p.log, msg)
}

func (p *streamingValidationHandlerConn) Send(msg any) error {
	if err := validateResponse(p.log, msg); err != nil {
		p.log.Error("failed sending message", logging.Error(err))
		return connect.NewError(connect.CodeInternal, eris.New("failed sending message"))
	}
	return p.StreamingHandlerConn.Send(msg)
}
