package elasticlog

import (
	"context"
	"fmt"
	"strings"

	"github.com/olivere/elastic"
	"gitlab.com/proemergotech/log-go"
)

type logger struct {
	log log.Logger
}

func NewLogger(log log.Logger) elastic.Logger {
	if !log.IsDebug(context.Background()) {
		return nil
	}

	return &logger{log: log}
}

func (el *logger) Printf(format string, v ...interface{}) {
	if el == nil {
		return
	}

	typ := "other"
	msg := fmt.Sprintf(format, v...)
	if strings.Contains(msg, "Accept: application/json") {
		typ = "request"
	} else if strings.Contains(msg, "Content-Type: application/json") {
		typ = "response"
	}

	el.log.Debug(context.Background(), "Elasticsearch out: "+typ, "message", fmt.Sprintf(format, v...))
}
