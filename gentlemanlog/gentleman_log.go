package gentlemanlog

import (
	"net/http"
	"net/http/httputil"

	gcontext "gopkg.in/h2non/gentleman.v2/context"
	"gopkg.in/h2non/gentleman.v2/plugin"

	"github.com/proemergotech/log/v3"
)

const maxBodyLength = 5000

type reqDumpKey struct{}
type resDumpKey struct{}

// Middleware return a middleware which will log additional data, if debug level is enabled in the logger.
// Request and response will be logged, based on the passed parameters.
// If the request/response body is more than 5000 bytes, it will be ignored.
func Middleware(l log.Logger, logRequest bool, logResponse bool) plugin.Plugin {
	end := func(gCtx *gcontext.Context, handler gcontext.Handler) {
		if !l.IsDebug(gCtx.Request.Context()) {
			handler.Next(gCtx)
			return
		}

		var reqDump []byte
		if rd := gCtx.Get(reqDumpKey{}); rd != nil {
			reqDump = rd.([]byte)
		}
		var resDump []byte
		if rd := gCtx.Get(resDumpKey{}); rd != nil {
			resDump = rd.([]byte)
		}

		req := gCtx.Request
		res := gCtx.Response
		err := gCtx.Error
		fields := make([]interface{}, 0)
		fields = request(fields, logRequest, req, reqDump)
		fields = response(fields, logResponse, res, resDump)

		message := "HTTP out: [" + req.Method + "] " + req.URL.String()
		if err != nil {
			fields = append(fields, "error", err)
			message = "error " + message + ": " + err.Error()
		}

		l.Debug(req.Context(), message, fields...)

		handler.Next(gCtx)
	}

	handlers := plugin.Handlers{
		"response": end,
		"error":    end,
	}

	if logRequest {
		handlers["before dial"] = func(gCtx *gcontext.Context, handler gcontext.Handler) {
			if !l.IsDebug(gCtx.Request.Context()) {
				handler.Next(gCtx)
				return
			}

			req := gCtx.Request
			reqDump, _ := httputil.DumpRequest(req, req.ContentLength < maxBodyLength)
			gCtx.Set(reqDumpKey{}, reqDump)

			handler.Next(gCtx)
		}
	}

	if logResponse {
		handlers["after dial"] = func(gCtx *gcontext.Context, handler gcontext.Handler) {
			if !l.IsDebug(gCtx.Request.Context()) {
				handler.Next(gCtx)
				return
			}

			res := gCtx.Response
			resDump, _ := httputil.DumpResponse(res, res.ContentLength < maxBodyLength)
			gCtx.Set(resDumpKey{}, resDump)

			handler.Next(gCtx)
		}
	}

	return &plugin.Layer{Handlers: handlers}
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
