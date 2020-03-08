package middleware

import (
	"net/http"
	"time"
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

func Combine(funcs ...middleware) middleware {
	return func(f handler) handler {
		for _, m := range funcs {
			f = m(f)
		}
		return f
	}
}

func HTTPResponseCodeMonitor(f handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		writer := &responseCodeMonitorWriter{
			w, http.StatusOK,
		}
		f(writer, r)
		writer.monitor()
	}
}

func HTTPResponseTimeMonitor(f handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		f(w, r)
		elapsed := time.Since(start)
		ResponseTime.Observe(float64(elapsed.Milliseconds()))
	}
}

var requests []time.Time

func HTTPRequestCountMonitor(f handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, time.Now())
		f(w, r)
	}
}

func HTTPRequestCounter() {
	for {
		var tmp5 []time.Time
		var tmp15 []time.Time
		var tmp60 []time.Time
		for _, t := range requests {
			if time.Since(t) <= time.Minute*5 {
				tmp5 = append(tmp5, t)
			}
			if time.Since(t) <= time.Minute*15 {
				tmp15 = append(tmp15, t)
			}
			if time.Since(t) <= time.Minute*60 {
				tmp60 = append(tmp60, t)
			}
		}
		RequestsLast5Min.Set(float64(len(tmp5)))
		RequestsLast15Min.Set(float64(len(tmp15)))
		RequestsLast60Min.Set(float64(len(tmp60)))
		requests = tmp60 //all requests are discarded when they are one hour old
		time.Sleep(time.Minute)
	}
}
