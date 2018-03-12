package zaplog

import (
	"context"
	"strconv"

	"gitlab.com/proemergotech/log-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	sugar     *zap.SugaredLogger
	ctxMapper log.ContextMapper
	debug     bool
}

type fields []interface{}

// NewLogger creates a new Logger backed by zap.
// This logger will check the debug level settings of the passed zap.Logger instance
// and handle some methods differently, based on that.
func NewLogger(zapLogger *zap.Logger, ctxLogger log.ContextMapper) log.Logger {
	return &logger{
		sugar:     zapLogger.Sugar(),
		ctxMapper: ctxLogger,
		debug:     zapLogger.Core().Enabled(zapcore.DebugLevel),
	}
}

func (l *logger) IsDebug(ctx context.Context) bool {
	return l.debug
}

func (l *logger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, fields(keysAndValues).addAll(l.ctxMapper.Values(ctx)).processFields()...)
}

func (l *logger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, fields(keysAndValues).addAll(l.ctxMapper.Values(ctx)).processFields()...)
}

func (l *logger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, fields(keysAndValues).addAll(l.ctxMapper.Values(ctx)).processFields()...)
}

func (l *logger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, fields(keysAndValues).addAll(l.ctxMapper.Values(ctx)).processFields()...)
}

func (l *logger) Panic(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.sugar.Panicw(msg, fields(keysAndValues).addAll(l.ctxMapper.Values(ctx)).processFields()...)
}

func (l *logger) Dump(msg string, v ...interface{}) {
	if !l.debug {
		return
	}

	args := make([]interface{}, 0, len(v)*2)
	for i, arg := range v {
		args = append(args, "arg"+strconv.Itoa(i), arg)
	}

	l.sugar.Debugw(msg, fields(args).processFields()...)
}

func (f fields) processFields() fields {
	for _, v := range f {
		errI := v
		if zField, ok := v.(zapcore.Field); ok && zField.Type == zapcore.ErrorType {
			errI = zField.Interface
		}

		if err, ok := errI.(error); ok {
			f = append(f, errorFields(err)...)
		}
	}

	return f
}

func (f fields) addAll(m map[string]string) fields {
	for k, v := range m {
		f = append(f, zap.String(k, v))
	}

	return f
}

func errorFields(err error) []interface{} {
	type causer interface {
		Cause() error
	}

	type fielder interface {
		Fields() []interface{}
	}

	var fields []interface{}
	for err != nil {
		if fErr, ok := err.(fielder); ok {
			fields = append(fields, fErr.Fields()...)
		}

		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}

	return fields
}
