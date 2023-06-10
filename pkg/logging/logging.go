package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is an interface aiming to make the type of the underlying logger redundant
type Logger interface {
	Debug(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Level() Level
	Log(lvl Level, msg string, fields ...Field)
	Named(s string) *zap.Logger
	Panic(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	With(fields ...Field) *zap.Logger
}

type loggerCtxKeyType int
type Level = zapcore.Level
type Field = zap.Field

const (
	loggerCtxKey loggerCtxKeyType = iota

	DebugLevel Level = zapcore.DebugLevel
	InfoLevel  Level = zapcore.InfoLevel
	WarnLevel  Level = zapcore.WarnLevel
	ErrorLevel Level = zapcore.ErrorLevel
	PanicLevel Level = zapcore.PanicLevel
	FatalLevel Level = zapcore.FatalLevel
)

var logger *zap.Logger

func init() {
	if l, err := zap.NewProduction(); err == nil {
		logger = l
	} else {
		panic(err)
	}
}

func GetLogger() Logger {
	return logger
}

// IntoContext packages a logger into a given context
func IntoContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

// FromContext extracts a logger from a context or defaults to the default logger if none is present
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return logger
	}
	if ctxLogger, ok := ctx.Value(loggerCtxKey).(Logger); ok {
		return ctxLogger
	}
	return logger
}

var Any = zap.Any
var Binary = zap.Binary
var Bool = zap.Bool
var Complex128 = zap.Complex128
var Complex64 = zap.Complex64
var Duration = zap.Duration
var Error = zap.Error
var Float64 = zap.Float64
var Float32 = zap.Float32
var Int = zap.Int
var Int64 = zap.Int64
var Int32 = zap.Int32
var Int16 = zap.Int16
var Int8 = zap.Int8
var String = zap.String
var Time = zap.Time
var Uint = zap.Uint
var Uint64 = zap.Uint64
var Uint32 = zap.Uint32
var Uint16 = zap.Uint16
var Uint8 = zap.Uint8

func Trace(traceId string) Field {
	return String("traceId", traceId)
}
