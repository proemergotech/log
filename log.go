package log

import (
	"context"
	"unicode/utf8"
)

const (
	AppName    = "app_name"
	AppVersion = "app_version"
)

// Logger is an interface that wraps the most common logging methods.
// Correlation related information like the Correlation-Id and Workflow-Id
// will be extracted from the passed Context.
type Logger interface {
	Debug(ctx context.Context, msg string, keysAndValues ...interface{})
	Info(ctx context.Context, msg string, keysAndValues ...interface{})
	Warn(ctx context.Context, msg string, keysAndValues ...interface{})
	Error(ctx context.Context, msg string, keysAndValues ...interface{})
	Panic(ctx context.Context, msg string, keysAndValues ...interface{})
	IsDebug(ctx context.Context) bool
	// Dump should be used only during development and should not stay in production code.
	Dump(msg string, v ...interface{})
}

// ContextMapper is an interface that returns Values which should be included in the log fields.
type ContextMapper interface {
	Values(ctx context.Context) map[string]string
}

func Truncate(str string, length int, concat string) string {
	if len(str) <= length {
		return str
	}

	finalLength := length - len(concat)
	headLength := finalLength/2 + finalLength%2
	tailLength := finalLength / 2

	for !utf8.RuneStart(str[headLength]) {
		headLength--
	}
	for !utf8.RuneStart(str[tailLength]) {
		tailLength++
	}

	head := str[0:headLength]
	tail := str[len(str)-tailLength:]

	return head + concat + tail
}
