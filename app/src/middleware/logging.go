package middleware

import (
	"net/http"

	"github.com/matt035343/devops/app/src/log"
)

//HTTPErrorLoggerMiddleware Logs debug info about requests and warns about HTTP 500 responses.
func HTTPErrorLoggerMiddleware(f Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("New request from %s to %s", r.RemoteAddr, r.URL)
		writer := &responseCodeMonitorWriter{
			w, http.StatusOK,
		}
		f(writer, r)
		if writer.statusCode < 500 {
			log.Debug("Request responded with status HTTP %d", writer.statusCode)
		} else {
			log.Warning("Request responded with status HTTP %d", writer.statusCode)
		}
	}
}
