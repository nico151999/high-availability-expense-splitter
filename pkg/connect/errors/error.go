package errors

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func NewErrorWithDetails(ctx context.Context, code connect.Code, msg string, details []protoreflect.ProtoMessage) *connect.Error {
	log := logging.FromContext(ctx)

	conErr := connect.NewError(
		code,
		eris.New(msg),
	)
	for _, detail := range details {
		if detail, detailErr := connect.NewErrorDetail(detail); detailErr == nil {
			conErr.AddDetail(detail)
		} else {
			log.Error("unexpected error creating error metadata", logging.Error(conErr))
			return connect.NewError(code, eris.New("failed creating error details"))
		}
	}
	return conErr
}

// DetailsFromError returns the details from index start to end (excluding end) of a connect error. If you want all details to be returned pass end = -1.
func DetailsFromError(ctx context.Context, err error, start, end int) ([]protoreflect.ProtoMessage, error) {
	log := logging.FromContext(ctx)

	var connectErr *connect.Error
	if !eris.As(err, &connectErr) {
		log.Error("the error was expetced to be a connect error but it is not")
		return nil, eris.New("the error is not a connect error")
	}
	details := connectErr.Details()
	size := len(details)
	if start < 0 || end > size {
		log.Error("an invalid range was passed for error details to be extracted", logging.Int("start", start), logging.Int("end", end))
		return nil, eris.New("the range must be within the detail size")
	}
	if end == -1 {
		end = size
	} else {
		size = end - start
	}
	res := make([]protoreflect.ProtoMessage, size)
	for i := start; i < end; i++ {
		var valueErr error
		res[i], valueErr = details[i].Value()
		if valueErr != nil {
			log.Error("failed getting detail value", logging.Error(err))
			return nil, eris.Wrap(valueErr, "failed getting detail value")
		}
	}
	return res, nil
}

func TypedDetailsFromError[T protoreflect.ProtoMessage](ctx context.Context, err error, start, end int) ([]T, error) {
	log := logging.FromContext(ctx)

	details, err := DetailsFromError(ctx, err, 0, -1)
	if err != nil {
		return nil, eris.Wrap(err, "could not get details")
	}
	res := make([]T, len(details))
	for i, detail := range details {
		if detail, ok := detail.(T); ok {
			res[i] = detail
		} else {
			log.Error("an error detail is not of the specified type", logging.Int("index", i))
			return nil, eris.Errorf("the %d. error detail is not of the specified type", i)
		}
	}
	return res, nil
}

func FirstTypedDetailFromError[T protoreflect.ProtoMessage](ctx context.Context, err error) (T, error) {
	details, err := TypedDetailsFromError[T](ctx, err, 0, 1)
	if err != nil {
		var zeroVal T
		return zeroVal, err
	}
	return details[0], nil
}
