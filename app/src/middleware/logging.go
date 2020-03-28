package middleware

import (
	"net/http"

	"github.com/matt035343/devops/app/src/log"
)

//HTTPErrorLoggerMiddleware Logs debug info about requests and warns about HTTP 500 responses.
func HTTPErrorLoggerMiddleware(f Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("New request from %s to %s", r.RemoteAddr, r.URL)
		f(w, r)
		if r.Response.StatusCode < 500 {
			log.Debug("Request responded with status HTTP %d", r.Response.StatusCode)
		} else {
			log.Warning("Request responded with status HTTP %d", r.Response.StatusCode)
		}
	}
}
