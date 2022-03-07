package httplog

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/proemergotech/log/v3"
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
		var content []byte
		var gzipEncoded bool
		for _, value := range req.Header.Values("Content-Encoding") {
			if value == "gzip" {
				gzipEncoded = true
				break
			}
		}

		if gzipEncoded {
			content, _ = ioutil.ReadAll(req.Body)
			req.Body = decompressGzipContent(ioutil.NopCloser(bytes.NewReader(content)))
		}
		reqDump, _ = httputil.DumpRequest(req, true)

		fields = append(fields, "request", log.Truncate(string(reqDump), maxBodyLength, "\n...\n"))

		if gzipEncoded {
			req.Body = ioutil.NopCloser(bytes.NewReader(content))
		}
	}

	res, err := t.inner.RoundTrip(req)

	var resDump []byte
	var responseLength = -1
	if err == nil && t.logResponse && t.logger.IsDebug(req.Context()) {
		resDump, _ = httputil.DumpResponse(res, true)
		responseLength = len(resDump)
		fields = append(fields, "response", log.Truncate(string(resDump), maxBodyLength, "\n...\n"))
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

func decompressGzipContent(body io.ReadCloser) io.ReadCloser {
	var buf bytes.Buffer
	gzr, err := gzip.NewReader(body)
	if err != nil {
		return body
	}
	_, err = io.Copy(&buf, gzr) // #nosec
	if err != nil {
		return body
	}

	if err := gzr.Close(); err != nil {
		return body
	}

	return ioutil.NopCloser(&buf)
}
