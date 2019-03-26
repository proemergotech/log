package httplog

import (
	"net/http"
	"net/http/httputil"

	"gitlab.com/proemergotech/log-go"
)

const maxBodyLength = 5000

type loggingTransport struct {
	prefix      string
	inner       http.RoundTripper
	logger      log.Logger
	logRequest  bool
	logResponse bool
}

func NewLoggingTransport(transport http.RoundTripper, logger log.Logger, prefix string, logRequest bool, logResponse bool) http.RoundTripper {
	return &loggingTransport{
		prefix:      prefix,
		inner:       transport,
		logger:      logger,
		logRequest:  logRequest,
		logResponse: logResponse,
	}
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fields := make([]interface{}, 0)
	var reqDump []byte
	if t.logRequest && t.logger.IsDebug(req.Context()) {
		reqDump, _ = httputil.DumpRequest(req, true)

		fields = append(fields, "request", string(limitBody(reqDump)))
	}

	res, err := t.inner.RoundTrip(req)

	var resDump []byte
	var responseLength = -1
	if err == nil && t.logResponse && t.logger.IsDebug(req.Context()) {
		resDump, _ = httputil.DumpResponse(res, true)
		responseLength = len(resDump)
		fields = append(fields, "response", string(limitBody(resDump)))
	}

	fields = append(fields, "request-length", req.ContentLength)
	// res.ContentLength is not set
	fields = append(fields, "response-length", responseLength)

	message := "HTTP out: [" + req.Method + "] " + req.URL.String()
	if err != nil {
		fields = append(fields, "error", err)
		message = "error " + message + ": " + err.Error()
	}

	t.logger.Debug(req.Context(), t.prefix+message, fields...)

	return res, err
}

func limitBody(body []byte) []byte {
	if len(body) <= maxBodyLength {
		return body
	}

	concat := "\n...\n"
	headTailLength := (maxBodyLength - len(concat)) / 2
	head := body[0:headTailLength]
	tail := body[len(body)-headTailLength:]
	body = make([]byte, 0, maxBodyLength)
	body = append(body, head...)
	body = append(body, concat...)
	body = append(body, tail...)

	return body
}
