package ginlog

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
	"gitlab.com/proemergotech/log-go"
)

const maxBodyLength = 5000

type recordingWriter struct {
	gin.ResponseWriter
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
	cnt, err := cw.ResponseWriter.WriteString(str)
	if err != nil {
		return cnt, err
	}

	cw.body.WriteString(str)

	return cnt, err
}

// ErrorMiddleware return a middleware which will log all the Errors in *gin.Context depends on the status code.
// When "status >= 400 && status < 500" it will be logged as a warn, otherwise error.
func ErrorMiddleware(l log.Logger) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		gCtx.Next()

		if len(gCtx.Errors) > 0 {
			status := gCtx.Writer.Status()

			for _, e := range gCtx.Errors {
				err := e.Err
				if err != nil {
					if status >= 400 && status < 500 {
						l.Warn(gCtx.Request.Context(), err.Error(), "error", err)
					} else {
						l.Error(gCtx.Request.Context(), err.Error(), "error", err)
					}
				}
			}
		}
	}
}

// DebugMiddleware return a middleware which will log additional data, if debug level is enabled in the logger.
// Request and response will be logged, based on the passed parameters.
// If the request/response body is more than 5000 bytes, it will be ignored.
func DebugMiddleware(l log.Logger, logRequest bool, logResponse bool) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		if !l.IsDebug(gCtx.Request.Context()) {
			gCtx.Next()
			return
		}

		var reqDump []byte
		var res *http.Response
		var resDump []byte

		req := gCtx.Request
		if logRequest {
			reqDump, _ = httputil.DumpRequest(req, req.ContentLength < maxBodyLength)
		}
		if logResponse {
			rw := &recordingWriter{ResponseWriter: gCtx.Writer, body: new(bytes.Buffer)}
			gCtx.Writer = rw
		}

		gCtx.Next()

		res = &http.Response{
			ProtoMajor:    req.ProtoMajor,
			ProtoMinor:    req.ProtoMinor,
			Proto:         req.Proto,
			ContentLength: int64(gCtx.Writer.Size()),
			Status:        http.StatusText(gCtx.Writer.Status()),
			StatusCode:    gCtx.Writer.Status(),
			Header:        gCtx.Writer.Header(),
			Request:       gCtx.Request,
		}

		if logResponse {
			res.Body = ioutil.NopCloser(gCtx.Writer.(*recordingWriter).body)
			resDump, _ = httputil.DumpResponse(res, res.ContentLength < maxBodyLength)
		}

		var err error
		if len(gCtx.Errors) > 0 {
			err = gCtx.Errors[0].Err
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
