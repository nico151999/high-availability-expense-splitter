package otel

import (
	"context"
	"math"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func NewOtelInterceptorFunc(ctx context.Context) logging.InterceptorFunc {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	return func(next func(lvl logging.Level, msg string, fields ...logging.Field), lvl logging.Level, msg string, fields ...logging.Field) {
		fields = append(
			fields,
			logging.String("traceId", spanCtx.TraceID().String()),
			logging.String("spanId", spanCtx.SpanID().String()))
		registerSpanEvent(span, lvl, msg, fields...)
		next(lvl, msg, fields...)
	}
}

func registerSpanEvent(span trace.Span, lvl logging.Level, msg string, fields ...logging.Field) {
	switch lvl {
	// we automatically set the span to error if an erroneous log is printed but
	// we do not automatically set it to something else otherwise.
	case logging.ErrorLevel, logging.FatalLevel, logging.PanicLevel:
		span.SetStatus(codes.Error, msg)
	}
	attributes := []attribute.KeyValue{
		attribute.String("level", lvl.String()),
		attribute.String("message", msg),
	}
	for _, field := range fields {
		switch field.Type {
		case logging.BoolType:
			attributes = append(attributes, attribute.Bool(field.Key, field.Integer == 1))
		case logging.ErrorType:
			err := field.Interface.(error)
			attributes = append(attributes, attribute.String(field.Key, err.Error()))
			span.RecordError(err)
		case logging.Float64Type:
			attributes = append(attributes, attribute.Float64(field.Key, math.Float64frombits(uint64(field.Integer))))
		case logging.Float32Type:
			attributes = append(attributes, attribute.Float64(field.Key, float64(math.Float32frombits(uint32(field.Integer)))))
		case logging.Int64Type:
			attributes = append(attributes, attribute.Int64(field.Key, field.Integer))
		case logging.Int32Type:
			attributes = append(attributes, attribute.Int64(field.Key, field.Integer))
		case logging.Int16Type:
			attributes = append(attributes, attribute.Int64(field.Key, field.Integer))
		case logging.Int8Type:
			attributes = append(attributes, attribute.Int64(field.Key, field.Integer))
		case logging.StringType:
			attributes = append(attributes, attribute.String(field.Key, field.String))
		default:
			attributes = append(attributes, attribute.String(field.Key, "UNSUPPORTED FIELD TYPE"))
		}
	}
	span.AddEvent("log", trace.WithAttributes(attributes...))
}
