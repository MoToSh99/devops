package utils

import (
	"crypto/md5" /* #nosec G501 */
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/matt035343/devops/app/src/log"
	"github.com/matt035343/devops/app/src/middleware"
)

// GravatarURL Generates Gravatar image URL from email and wanted image height in pixels (results in square-sized image)
func GravatarURL(email string, size int) string {
	cleanedEmail := strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(cleanedEmail)) /* #nosec G401 */
	hex := hex.EncodeToString(hash[:])
	return fmt.Sprintf("http://www.gravatar.com/avatar/%s?d=identicon&s=%d", hex, size)
}

// ExternalMonitor Monitors another Minitwit server located at the given API URL
func ExternalMonitor(url string) {
	log.Info("Connecting to external server on %s", url)
	for {
		t := time.Now()
		resp, err := http.Get(url + "/latest") /* #nosec G107 */
		if err != nil {
			middleware.ExternalMonitorUnssuccessfulRequests.Inc()
			log.WarningErr("Could not connect to external server to be monitored", err)
		} else if resp.StatusCode != 200 && resp.StatusCode != 204 {
			middleware.ExternalMonitorUnssuccessfulRequests.Inc()
			log.Warning("Could not connect to external server to be monitored, HTTP %d", resp.StatusCode)
		} else {
			middleware.ExternalMonitorResponseTime.Observe(float64(time.Since(t).Milliseconds()))
		}
		time.Sleep(time.Minute)
	}
}
