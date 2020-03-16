package middleware

import (
	"net/http"
)

type handler func(http.ResponseWriter, *http.Request)
type middleware func(handler) handler

type responseCodeMonitorWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *responseCodeMonitorWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *responseCodeMonitorWriter) monitor() {
	HTTPResponses.Observe(float64(lrw.statusCode))
}

//Combine Combines 1 or more middlewares, applied in the given order.
func Combine(funcs ...middleware) middleware {
	return func(f handler) handler {
		for _, m := range funcs {
			f = m(f)
		}
		return f
	}
}
