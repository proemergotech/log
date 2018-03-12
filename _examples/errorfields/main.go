package main

import (
	"fmt"

	"gitlab.com/proemergotech/log-go"
	"gitlab.com/proemergotech/log-go/zaplog"
	trace "gitlab.com/proemergotech/trace-go"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"context"
	"io"

	"github.com/pkg/errors"
)

func init() {
	err := zap.RegisterEncoder(zaplog.EncoderType, zaplog.NewEncoder([]string{
		trace.CorrelationIDField,
		trace.WorkflowIDField,
		log.AppName,
		log.AppVersion,
	}))
	if err != nil {
		panic(fmt.Sprintf("couldn't create logger: %v", err))
	}

	zapConf := zap.NewProductionConfig()
	zapConf.Encoding = zaplog.EncoderType

	logLvl := new(zapcore.Level)
	logLvl.Set("debug")
	zapConf.Level = zap.NewAtomicLevelAt(*logLvl)

	zapLog, err := zapConf.Build()
	if err != nil {
		panic(fmt.Sprintf("couldn't create logger: %v", err))
	}
	zapLog = zapLog.With(
		zap.String(log.AppName, "error example"),
		zap.String(log.AppVersion, "dev"),
	)

	logger := zaplog.NewLogger(zapLog, trace.Mapper())
	log.SetGlobalLogger(logger)
}

type withFields struct {
	cause  error
	fields []interface{}
}

func WithFields(err error, fields ...interface{}) error {
	if len(fields) == 0 {
		return err
	}

	return &withFields{
		cause:  err,
		fields: fields,
	}
}

func (e *withFields) Error() string {
	return e.cause.Error()
}

func (e *withFields) Fields() []interface{} {
	return e.fields
}

func (e *withFields) Cause() error {
	return e.cause
}

func (e *withFields) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", e.Cause())
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, e.Error())
	}
}

func main() {
	log.Info(context.Background(), "hello world", "world", "earth")

	err := errors.New("this is bad")
	err = WithFields(err, "big", "boom")
	log.Error(context.Background(), "goodbye world", "error", err, "world", "earth")
}
