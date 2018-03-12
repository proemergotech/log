package log

import (
	"context"
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
