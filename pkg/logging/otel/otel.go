package otel

import (
	"context"
	"math"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type OtelLogger interface {
	Debug(msg string, fields ...logging.Field)
	Error(msg string, fields ...logging.Field)
	Fatal(msg string, fields ...logging.Field)
	Info(msg string, fields ...logging.Field)
	Panic(msg string, fields ...logging.Field)
	Warn(msg string, fields ...logging.Field)
}

type otelLogger struct {
	logger logging.Logger
	span   trace.Span
}

// NewOtelLogger creates a new instance of an otel logger out of the span from a context and a normal logger
func NewOtelLogger(ctx context.Context, logger logging.Logger) OtelLogger {
	span := trace.SpanFromContext(ctx).SpanContext()

	return &otelLogger{
		logger.With(
			logging.String("traceId", span.TraceID().String()),
			logging.String("spanId", span.SpanID().String())),
		trace.SpanFromContext(ctx),
	}
}

// NewOtelLogger creates a new instance of an otel logger out of both the normal logger and the span from a context
func NewOtelLoggerFromContext(ctx context.Context) OtelLogger {
	return NewOtelLogger(ctx, logging.FromContext(ctx))
}

func (l *otelLogger) Debug(msg string, fields ...logging.Field) {
	registerSpanEvent(l.span, "debug", msg, fields...)
	l.logger.Debug(msg, fields...)
}

func (l *otelLogger) Error(msg string, fields ...logging.Field) {
	l.span.SetStatus(codes.Error, msg)
	registerSpanEvent(l.span, "error", msg, fields...)
	l.logger.Error(msg, fields...)
}

func (l *otelLogger) Fatal(msg string, fields ...logging.Field) {
	l.span.SetStatus(codes.Error, msg)
	registerSpanEvent(l.span, "fatal", msg, fields...)
	l.logger.Fatal(msg, fields...)
}

func (l *otelLogger) Info(msg string, fields ...logging.Field) {
	registerSpanEvent(l.span, "info", msg, fields...)
	l.logger.Info(msg, fields...)
}

func (l *otelLogger) Panic(msg string, fields ...logging.Field) {
	l.span.SetStatus(codes.Error, msg)
	registerSpanEvent(l.span, "panic", msg, fields...)
	l.logger.Panic(msg, fields...)
}

func (l *otelLogger) Warn(msg string, fields ...logging.Field) {
	registerSpanEvent(l.span, "warn", msg, fields...)
	l.logger.Warn(msg, fields...)
}

func registerSpanEvent(span trace.Span, level string, msg string, fields ...logging.Field) {
	attributes := []attribute.KeyValue{
		attribute.String("level", level),
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
