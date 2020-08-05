package log

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
)

type ThrottleLogger struct {
	logger   Logger
	interval time.Duration
	logs     *sync.Map

	wg                *sync.WaitGroup
	closeCh, closedCh chan struct{}
}

type throttleKey struct {
	level int
	msg   string
}

type throttleLog struct {
	ctx           context.Context
	count         *int32
	keysAndValues []interface{}
}

func Throttle(logger Logger, interval time.Duration) (*ThrottleLogger, chan struct{}) {
	closeCh := make(chan struct{})
	closedCh := make(chan struct{})
	wg := new(sync.WaitGroup)

	t := &ThrottleLogger{
		logger:   logger,
		interval: interval,
		logs:     &sync.Map{},
		wg:       wg,
		closeCh:  closeCh,
		closedCh: closedCh,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.Flush()
			case <-closeCh:
				return
			}
		}
	}()

	return t, closedCh
}

func (t *ThrottleLogger) Stop() {
	t.close()
}

func (t *ThrottleLogger) Flush() {
	t.logs.Range(func(key, value interface{}) bool {
		tKey := key.(throttleKey)
		tLog := value.(*throttleLog)

		if tLog == nil {
			return true
		}

		logCount := atomic.LoadInt32(tLog.count)
		if logCount == 0 {
			t.logs.Delete(key)
			return true
		}

		keysAndValues := tLog.keysAndValues
		if logCount > 1 {
			keysAndValues = append(keysAndValues, "times", logCount)
		}

		t.logMessage(tLog.ctx, tKey.level, tKey.msg, keysAndValues...)

		atomic.AddInt32(tLog.count, -logCount)

		return true
	})
}

func (t *ThrottleLogger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	t.throttleLogMessage(ctx, DebugLevel, msg, keysAndValues...)
}

func (t *ThrottleLogger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	t.throttleLogMessage(ctx, InfoLevel, msg, keysAndValues...)
}

func (t *ThrottleLogger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	t.throttleLogMessage(ctx, WarnLevel, msg, keysAndValues...)
}

func (t *ThrottleLogger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	t.throttleLogMessage(ctx, ErrorLevel, msg, keysAndValues...)
}

func (t *ThrottleLogger) Panic(ctx context.Context, msg string, keysAndValues ...interface{}) {
	t.throttleLogMessage(ctx, PanicLevel, msg, keysAndValues...)
}

func (t *ThrottleLogger) IsDebug(ctx context.Context) bool {
	return t.logger.IsDebug(ctx)
}

// Dump should be used only during development and should not stay in production code.
func (t *ThrottleLogger) Dump(msg string, v ...interface{}) {
	t.logger.Dump(msg, v...)
}

func (t *ThrottleLogger) throttleLogMessage(ctx context.Context, logLevel int, msg string, keysAndValues ...interface{}) {
	log, ok := t.logs.LoadOrStore(throttleKey{level: logLevel, msg: msg}, &throttleLog{ctx: ctx, count: new(int32), keysAndValues: keysAndValues})
	if !ok {
		// first time, call log
		t.logMessage(ctx, logLevel, msg, keysAndValues...)
	} else {
		atomic.AddInt32(log.(*throttleLog).count, 1)
	}
}

func (t *ThrottleLogger) logMessage(ctx context.Context, logLevel int, msg string, keysAndValues ...interface{}) {
	var logFn func(context.Context, string, ...interface{})

	switch logLevel {
	case DebugLevel:
		logFn = t.logger.Debug
	case InfoLevel:
		logFn = t.logger.Info
	case WarnLevel:
		logFn = t.logger.Warn
	case ErrorLevel:
		logFn = t.logger.Error
	case PanicLevel:
		logFn = t.logger.Panic
	default:
		return
	}

	logFn(ctx, msg, keysAndValues...)
}

func (t *ThrottleLogger) close() {
	defer func() {
		t.wg.Wait()
		if t.closedCh != nil {
			close(t.closedCh)
			t.closedCh = nil
		}
	}()

	if t.closeCh == nil {
		return
	}

	close(t.closeCh)
	t.closeCh = nil
}
