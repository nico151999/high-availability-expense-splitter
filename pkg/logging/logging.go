package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = (*logger)(nil)

// Logger is an interface aiming to make the type of the underlying logger redundant
type Logger interface {
	Debug(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Level() Level
	Log(lvl Level, msg string, fields ...Field)
	Named(s string) Logger
	NewNamed(s string) Logger
	Panic(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	With(fields ...Field) Logger
}

type logger struct {
	*zap.Logger
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

var globalLogger *logger

func init() {
	if l, err := zap.NewProduction(); err == nil {
		globalLogger = &logger{l}
	} else {
		panic(err)
	}
}

func (l *logger) Named(s string) Logger {
	return &logger{
		l.Logger.Named(s),
	}
}

func (l *logger) NewNamed(s string) Logger {
	n := *l.Logger
	return &logger{
		n.Named(s),
	}
}

func (l *logger) With(fields ...Field) Logger {
	return &logger{
		l.Logger.With(fields...),
	}
}

func GetLogger() Logger {
	return globalLogger
}

// IntoContext packages a logger into a given context
func IntoContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

// FromContext extracts a logger from a context or defaults to the default logger if none is present
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return globalLogger
	}
	if ctxLogger, ok := ctx.Value(loggerCtxKey).(*logger); ok {
		return ctxLogger
	}
	return globalLogger
}

var Any = zap.Any
var Binary = zap.Binary
var BinaryType = zapcore.BinaryType
var Bool = zap.Bool
var BoolType = zapcore.BoolType
var Complex128 = zap.Complex128
var Complex128Type = zapcore.Complex128Type
var Complex64 = zap.Complex64
var Complex64Type = zapcore.Complex64Type
var Duration = zap.Duration
var DurationType = zapcore.DurationType
var Error = zap.Error
var ErrorType = zapcore.ErrorType
var Float64 = zap.Float64
var Float64Type = zapcore.Float64Type
var Float32 = zap.Float32
var Float32Type = zapcore.Float32Type
var Int = zap.Int
var Int64 = zap.Int64
var Int64Type = zapcore.Int64Type
var Int32 = zap.Int32
var Int32Type = zapcore.Int32Type
var Int16 = zap.Int16
var Int16Type = zapcore.Int16Type
var Int8 = zap.Int8
var Int8Type = zapcore.Int8Type
var String = zap.String
var StringType = zapcore.StringType
var Time = zap.Time
var TimeType = zapcore.TimeType
var Uint = zap.Uint
var Uint64 = zap.Uint64
var Uint64Type = zapcore.Uint64Type
var Uint32 = zap.Uint32
var Uint32Type = zapcore.Uint32Type
var Uint16 = zap.Uint16
var Uint16Type = zapcore.Uint16Type
var Uint8 = zap.Uint8
var Uint8Type = zapcore.Uint8Type
