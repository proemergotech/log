package elasticlog

import (
	"context"
	"fmt"

	"github.com/olivere/elastic"
	"gitlab.com/proemergotech/log-go"
)

type errorLogger struct {
	log log.Logger
}

func NewErrorLogger(log log.Logger) elastic.Logger {
	return &errorLogger{log: log}
}

func (el *errorLogger) Printf(format string, v ...interface{}) {
	el.log.Error(context.Background(), "Elasticsearch error", "message", fmt.Sprintf(format, v...))
}
