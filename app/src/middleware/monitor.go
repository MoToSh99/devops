package middleware

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//HTTPResponses A Prometheus histogram for recording HTTP response code.
var HTTPResponses = promauto.NewHistogram(prometheus.HistogramOpts{
	Subsystem: "minitwit",
	Name:      "http_responses",
	Buckets: []float64{
		100.0, 101.0, 102.0, 103.0,
		200.0, 201.0, 202.0, 203.0, 204.0, 205.0, 206.0, 207.0, 208.0, 226.0,
		300.0, 301.0, 302.0, 303.0, 304.0, 305.0, 306.0, 307.0, 308.0,
		400.0, 401.0, 402.0, 403.0, 404.0, 405.0, 406.0, 407.0, 408.0, 409.0, 410.0, 411.0, 412.0, 413.0, 414.0, 415.0, 416.0, 417.0, 418.0, 421.0, 422.0, 423.0, 424.0, 426.0, 428.0, 429.0, 431.0, 444.0, 451.0, 499.0,
		500.0, 501.0, 502.0, 503.0, 504.0, 505.0, 506.0, 507.0, 508.0, 510.0, 511.0, 599.0,
	},
})

//ResponseTime A Prometheus histogram for recording request response times.
var ResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
	Subsystem: "minitwit",
	Name:      "response_time_ms",
	Help:      "A histogram of the response time of all request coming into the website.",
	Buckets: []float64{
		0.1, 1.0, 5.0, 10.0, 15.0, 20.0, 30.0, 40.0, 50.0, 75.0, 100.0, 150.0, 200.0, 300.0, 400.0, 500.0, 750.0, 1000.0, 1500.0, 2000.0, 5000.0, 10000.0,
	},
})

//RequestsLast5Min A Prometheus gauge for recoding amount requests last 5 min.
var RequestsLast5Min = promauto.NewGauge(prometheus.GaugeOpts{
	Subsystem: "minitwit",
	Name:      "requests_last_5_min",
	Help:      "The number of requests received by the website within the last 5 minutes",
})

//RequestsLast15Min A Prometheus gauge for recoding amount requests last 15 min.
var RequestsLast15Min = promauto.NewGauge(prometheus.GaugeOpts{
	Subsystem: "minitwit",
	Name:      "requests_last_15_min",
	Help:      "The number of requests received by the website within the last 15 minutes",
})

//RequestsLast60Min A Prometheus gauge for recoding amount requests last 60 min.
var RequestsLast60Min = promauto.NewGauge(prometheus.GaugeOpts{
	Subsystem: "minitwit",
	Name:      "requests_last_60_min",
	Help:      "The number of requests received by the website within the last 60 minutes",
})

//MessagesSent A Prometheus counter for recording the total amount of messages sent on the site.
var MessagesSent = promauto.NewCounter(prometheus.CounterOpts{
	Subsystem: "minitwit",
	Name:      "messages_sent",
	Help:      "The number of messages sent by users on the website.",
})

//UsersRegistered A Prometheus counter for recording the total amount of registered users on the site.
var UsersRegistered = promauto.NewCounter(prometheus.CounterOpts{
	Subsystem: "minitwit",
	Name:      "users_registered",
	Help:      "The number of users registered on the website.",
})

//UsersFollowed A Prometheus counter for recording the total amount of user follow actions on the site.
var UsersFollowed = promauto.NewCounter(prometheus.CounterOpts{
	Subsystem: "minitwit",
	Name:      "users_followed",
	Help:      "The number of times a user has followed another user. Note that follow, unfollow, follow counts twice.",
})

//UsersUnfollowed A Prometheus counter for recording the total amount of user unfollow actions on the site.
var UsersUnfollowed = promauto.NewCounter(prometheus.CounterOpts{
	Subsystem: "minitwit",
	Name:      "users_unfollowed",
	Help:      "The number of times a user has unfollowed another user. Note that unfollow, follow, unfollow counts twice.",
})

//HTTPResponseCodeMonitor Middleware for recording HTTP response codes.
func HTTPResponseCodeMonitor(f Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		writer := &responseCodeMonitorWriter{
			w, http.StatusOK,
		}
		f(writer, r)
		writer.monitor()
	}
}

//HTTPResponseTimeMonitor Middleware for recording reponse times.
func HTTPResponseTimeMonitor(f Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		f(w, r)
		elapsed := time.Since(start)
		ResponseTime.Observe(float64(elapsed.Milliseconds()))
	}
}

var requests []time.Time

//HTTPRequestCountMonitor Middleware for recording incoming requests.
func HTTPRequestCountMonitor(f Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, time.Now())
		f(w, r)
	}
}

//HTTPRequestCounter A blocking function that keeps track of requests recorded by HTTPRequestCountMonitor and reports them to Prometheus. Never returns!
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
