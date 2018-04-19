package log

import "context"

var global Logger

// SetGlobalLogger sets the Logger instance used in global scope.
// Must be called during initialization as early as possible.
// It's not thread safe.
func SetGlobalLogger(l Logger) {
	global = l
}

func Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	global.Debug(ctx, msg, keysAndValues...)
}

func Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	global.Info(ctx, msg, keysAndValues...)
}

func Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	global.Warn(ctx, msg, keysAndValues...)
}

func Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	global.Error(ctx, msg, keysAndValues...)
}

func Panic(ctx context.Context, msg string, keysAndValues ...interface{}) {
	global.Panic(ctx, msg, keysAndValues...)
}

func IsDebug(ctx context.Context) bool {
	return global.IsDebug(ctx)
}

// Dump should be used only during development and should not stay in production code.
func Dump(msg string, v ...interface{}) {
	global.Dump(msg, v...)
}
