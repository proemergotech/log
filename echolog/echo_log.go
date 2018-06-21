package echolog

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/labstack/echo"
	"gitlab.com/proemergotech/log-go"
)

const maxBodyLength = 5000

type recordingWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (cw *recordingWriter) Write(buf []byte) (int, error) {
	cnt, err := cw.ResponseWriter.Write(buf)
	if err != nil {
		return cnt, err
	}

	cw.body.Write(buf)

	return cnt, err
}

func (cw *recordingWriter) WriteString(str string) (int, error) {
	cnt, err := cw.ResponseWriter.Write([]byte(str))
	if err != nil {
		return cnt, err
	}

	cw.body.WriteString(str)

	return cnt, err
}

// ErrorMiddleware return a middleware which will log all the Errors in *echo.Context depending on the status code.
// When "status >= 400 && status < 500" it will be logged as a warn, otherwise as an error.
func ErrorMiddleware(l log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(eCtx echo.Context) error {
			err := next(eCtx)

			if err != nil {
				status := eCtx.Response().Status
				if status >= 400 && status < 500 {
					l.Warn(eCtx.Request().Context(), err.Error(), "error", err)
				} else {
					l.Error(eCtx.Request().Context(), err.Error(), "error", err)
				}
			}

			return err
		}
	}
}

// DebugMiddleware return a middleware which will log additional data, if debug level is enabled in the logger.
// Request and response will be logged, based on the passed parameters.
// If the request/response body is more than 5000 bytes, it will be ignored.
func DebugMiddleware(l log.Logger, logRequest bool, logResponse bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(eCtx echo.Context) error {
			if !l.IsDebug(eCtx.Request().Context()) {
				return next(eCtx)
			}

			var reqDump []byte
			var res *http.Response
			var resDump []byte

			req := eCtx.Request()
			if logRequest {
				reqDump, _ = httputil.DumpRequest(req, req.ContentLength < maxBodyLength)
			}
			if logResponse {
				rw := &recordingWriter{ResponseWriter: eCtx.Response().Writer, body: new(bytes.Buffer)}
				eCtx.Response().Writer = rw
			}
			rw := eCtx.Response().Writer

			err := next(eCtx)

			res = &http.Response{
				ProtoMajor:    req.ProtoMajor,
				ProtoMinor:    req.ProtoMinor,
				Proto:         req.Proto,
				ContentLength: int64(eCtx.Response().Size),
				Status:        http.StatusText(eCtx.Response().Status),
				StatusCode:    eCtx.Response().Status,
				Header:        eCtx.Response().Header(),
				Request:       eCtx.Request(),
			}

			if logResponse {
				res.Body = ioutil.NopCloser(rw.(*recordingWriter).body)
				resDump, _ = httputil.DumpResponse(res, res.ContentLength < maxBodyLength)
			}

			fields := make([]interface{}, 0)
			fields = request(fields, logRequest, req, reqDump)
			fields = response(fields, logResponse, res, resDump)

			message := "HTTP in: [" + req.Method + "] " + req.URL.String()
			if err != nil {
				fields = append(fields, "error", err)
				message = "error " + message + ": " + err.Error()
			}

			l.Debug(req.Context(), message, fields...)

			return err
		}
	}
}

func request(fields []interface{}, logRequest bool, req *http.Request, reqDump []byte) []interface{} {
	if logRequest {
		if reqDump == nil {
			reqDump, _ = httputil.DumpRequest(req, false)
		}

		fields = append(fields, "request", string(reqDump))
	}
	return append(fields, "request-length", req.ContentLength)
}

func response(fields []interface{}, logResponse bool, res *http.Response, resDump []byte) []interface{} {
	if logResponse {
		if resDump == nil {
			resDump, _ = httputil.DumpResponse(res, false)
		}
		fields = append(fields, "response", string(resDump))
	}
	return append(fields, "response-length", res.ContentLength)
}
