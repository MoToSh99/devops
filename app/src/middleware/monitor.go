package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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

var ResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
	Subsystem: "minitwit",
	Name:      "response_time_ms",
	Help:      "A histogram of the response time of all request coming into the website.",
})

var RequestsLast5Min = promauto.NewGauge(prometheus.GaugeOpts{
	Subsystem: "minitwit",
	Name:      "requests_last_5_min",
	Help:      "The number of requests received by the website within the last 5 minutes",
})

var RequestsLast15Min = promauto.NewGauge(prometheus.GaugeOpts{
	Subsystem: "minitwit",
	Name:      "requests_last_15_min",
	Help:      "The number of requests received by the website within the last 15 minutes",
})

var RequestsLast60Min = promauto.NewGauge(prometheus.GaugeOpts{
	Subsystem: "minitwit",
	Name:      "requests_last_60_min",
	Help:      "The number of requests received by the website within the last 60 minutes",
})

var MessagesSent = promauto.NewCounter(prometheus.CounterOpts{
	Subsystem: "minitwit",
	Name:      "messages_sent",
	Help:      "The number of messages sent by users on the website.",
})

var UsersRegistered = promauto.NewCounter(prometheus.CounterOpts{
	Subsystem: "minitwit",
	Name:      "users_registered",
	Help:      "The number of users registered on the website.",
})

var UsersFollowed = promauto.NewCounter(prometheus.CounterOpts{
	Subsystem: "minitwit",
	Name:      "musers_followed",
	Help:      "The number of times a user has followed another user. Note that follow, unfollow, follow counts twice.",
})

var UsersUnfollowed = promauto.NewCounter(prometheus.CounterOpts{
	Subsystem: "minitwit",
	Name:      "users_unfollowed",
	Help:      "The number of times a user has unfollowed another user. Note that unfollow, follow, unfollow counts twice.",
})
