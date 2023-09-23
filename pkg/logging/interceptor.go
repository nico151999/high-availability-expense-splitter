package logging

import "go.uber.org/zap"

type InterceptorFunc = func(next func(lvl Level, msg string, fields ...Field), lvl Level, msg string, fields ...Field)

type interceptor struct {
	logger *zap.Logger
	fn     InterceptorFunc
	next   *interceptor
}

func (i *interceptor) call(lvl Level, msg string, fields ...Field) {
	if i.next == nil {
		i.logger.Log(lvl, msg, fields...)
	} else {
		i.fn(i.next.call, lvl, msg, fields...)
	}
}

func (l *logger) WithInterceptors(interFuncs ...InterceptorFunc) Logger {
	n := *l.Logger
	if len(interFuncs) == 0 {
		return &logger{
			&n,
			nil,
		}
	}
	initialInterceptor := interceptor{
		logger: &n,
		fn:     interFuncs[0],
	}
	latestInterceptor := &initialInterceptor
	for i := 1; i < len(interFuncs); i++ {
		latestInterceptor.next = &interceptor{
			logger: &n,
			fn:     interFuncs[i],
		}
		latestInterceptor = latestInterceptor.next
	}
	return &logger{
		&n,
		&initialInterceptor,
	}
}

func (l *logger) Log(lvl Level, msg string, fields ...Field) {
	if l.initialInterceptor == nil {
		l.Logger.Log(lvl, msg, fields...)
	} else {
		l.initialInterceptor.call(lvl, msg, fields...)
	}
}

func (l *logger) Debug(msg string, fields ...Field) {
	l.Log(DebugLevel, msg, fields...)
}

func (l *logger) Error(msg string, fields ...Field) {
	l.Log(ErrorLevel, msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...Field) {
	l.Log(FatalLevel, msg, fields...)
}

func (l *logger) Info(msg string, fields ...Field) {
	l.Log(InfoLevel, msg, fields...)
}

func (l *logger) Panic(msg string, fields ...Field) {
	l.Log(PanicLevel, msg, fields...)
}

func (l *logger) Warn(msg string, fields ...Field) {
	l.Log(WarnLevel, msg, fields...)
}
