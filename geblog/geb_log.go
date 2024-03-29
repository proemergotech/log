package geblog

import (
	"github.com/proemergotech/geb-client/v2/geb"

	"github.com/proemergotech/log/v3"
)

// OnEventErrorMiddleware return a middleware which will log errors on ERROR level. s.
func OnEventErrorMiddleware(l log.Logger) geb.Middleware {
	return func(e *geb.Event, next func(*geb.Event) error) error {
		if err := next(e); err != nil {
			l.Error(e.Context(), err.Error(), "error", err)
		}

		return nil
	}
}

// OnEventDebugMiddleware return a middleware which will log additional data about event if debug level is enabled in the logger.
// Event header and body will be logged, based on the passed parameters.
func OnEventDebugMiddleware(l log.Logger, logEventBody bool) geb.Middleware {
	return func(e *geb.Event, next func(*geb.Event) error) error {
		err := next(e)

		logEvent(l, logEventBody, "GEB in:", e, err)

		return err
	}
}

// PublishMiddleware return a middleware which will log additional data about event if debug level is enabled in the logger.
// Event header and body will be logged, based on the passed parameters.
func PublishMiddleware(l log.Logger, logEventBody bool) geb.Middleware {
	return func(e *geb.Event, next func(*geb.Event) error) error {
		err := next(e)

		logEvent(l, logEventBody, "GEB out:", e, err)

		return err
	}
}

func logEvent(l log.Logger, logEventBody bool, prefix string, e *geb.Event, err error) {
	if !l.IsDebug(e.Context()) {
		return
	}

	f := make([]interface{}, 0)

	f = append(f,
		"event_name", e.EventName(),
		"event_headers", e.Headers(),
	)
	if logEventBody {
		var eventBody interface{}
		_ = e.Unmarshal(&eventBody)
		f = append(f, "event_body", eventBody)
	}

	message := prefix + " " + e.EventName()
	if err != nil {
		message = "error " + message + ": " + err.Error()
		f = append(f, "error", err)
	}

	l.Debug(e.Context(), message, f...)
}
