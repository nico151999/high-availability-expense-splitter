package grpc

import (
	"fmt"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"google.golang.org/grpc/grpclog"
)

type grpcLoggerV2 struct {
	logging.Logger
}

var _ grpclog.LoggerV2 = (*grpcLoggerV2)(nil)

const (
	grpcLvlInfo int = iota
	grpcLvlWarn
	grpcLvlError
	grpcLvlFatal
)

var (
	// grpcToZapLevel maps gRPC log levels to zap log levels.
	// See https://pkg.go.dev/go.uber.org/zap@v1.16.0/zapcore#Level
	grpcToZapLevel = map[int]logging.Level{
		grpcLvlInfo:  logging.InfoLevel,
		grpcLvlWarn:  logging.WarnLevel,
		grpcLvlError: logging.ErrorLevel,
		grpcLvlFatal: logging.FatalLevel,
	}
)

func NewGrpcLoggerV2(logger logging.Logger) grpclog.LoggerV2 {
	return &grpcLoggerV2{logger}
}

func (l *grpcLoggerV2) Info(args ...interface{}) {
	msg, fields := argsToFields(args...)
	l.Logger.Info(msg, fields...)
}

func (l *grpcLoggerV2) Infoln(args ...interface{}) {
	l.Info(sprintln(args))
}

func (l *grpcLoggerV2) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *grpcLoggerV2) Warning(args ...interface{}) {
	msg, fields := argsToFields(args...)
	l.Logger.Warn(msg, fields...)
}

func (l *grpcLoggerV2) Warningln(args ...interface{}) {
	l.Warning(sprintln(args))
}

func (l *grpcLoggerV2) Warningf(format string, args ...interface{}) {
	l.Warning(fmt.Sprintf(format, args...))
}

func (l *grpcLoggerV2) Error(args ...interface{}) {
	msg, fields := argsToFields(args...)
	l.Logger.Error(msg, fields...)
}

func (l *grpcLoggerV2) Errorln(args ...interface{}) {
	l.Error(sprintln(args))
}

func (l *grpcLoggerV2) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *grpcLoggerV2) Fatal(args ...interface{}) {
	msg, fields := argsToFields(args...)
	l.Logger.Fatal(msg, fields...)
}

func (l *grpcLoggerV2) Fatalln(args ...interface{}) {
	l.Fatal(sprintln(args))
}

func (l *grpcLoggerV2) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

func (l *grpcLoggerV2) V(lvl int) bool {
	return l.Logger.Level().Enabled(grpcToZapLevel[lvl])
}

func sprintln(args []interface{}) string {
	s := fmt.Sprintln(args...)
	// Drop the new line character added by Sprintln
	return s[:len(s)-1]
}

// argsToFields turns args into fields that can be passed to the structured logger. It falls back to a printed version.
func argsToFields(args ...interface{}) (string, []logging.Field) {
	fields := []logging.Field{}
	if len(args)%2 == 0 {
		return fmt.Sprint(args...), fields
	} else if msg, ok := args[0].(string); ok {
		var key string
		for i := 1; i < len(args); i++ {
			if i%2 == 1 {
				if arg, ok := args[i].(string); ok {
					key = arg
					continue
				} else {
					return fmt.Sprint(args...), fields
				}
			}
			fields = append(fields, logging.Any(key, args[i]))
		}
		return msg, fields
	} else {
		return fmt.Sprint(args...), fields
	}
}
