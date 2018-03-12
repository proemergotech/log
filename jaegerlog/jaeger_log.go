package jaegerlog

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"gitlab.com/proemergotech/log-go"
)

type JaegerLogger struct {
	logger log.Logger
}

// NewJaegerLogger return a logger which implements github.com/uber/jaeger-client-go.Logger
func NewJaegerLogger(l log.Logger) *JaegerLogger {
	return &JaegerLogger{logger: l}
}

func (jl *JaegerLogger) Error(msg string) {
	jl.logger.Error(context.Background(), msg, errors.New(msg))
}

func (jl *JaegerLogger) Infof(msg string, v ...interface{}) {
	args := make([]interface{}, 0, len(v)*2)
	for i, arg := range v {
		args = append(args, "arg"+strconv.Itoa(i), arg)
	}

	jl.logger.Info(context.Background(), msg, args...)
}
