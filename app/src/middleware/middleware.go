package middleware

import (
	"net/http"
)

//Handler A HTTP handler for Gorilla mux
type Handler func(http.ResponseWriter, *http.Request)

//Middleware A wrapper to do actions before and after handler
type Middleware func(Handler) Handler

type responseCodeMonitorWriter struct {
	http.ResponseWriter
	statusCode int
}

//WriteHeads Writes a HTTP status code to the header.
func (lrw *responseCodeMonitorWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *responseCodeMonitorWriter) monitor() {
	HTTPResponses.Observe(float64(lrw.statusCode))
}

//Combine Combines 1 or more middlewares, applied in the given order.
func Combine(funcs ...Middleware) Middleware {
	return func(f Handler) Handler {
		for _, m := range funcs {
			f = m(f)
		}
		return f
	}
}

//Unit A unit middleware that does nothing
func Unit(f Handler) Handler {
	return f
}
