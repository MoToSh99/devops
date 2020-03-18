package utils

import (
	"crypto/md5" /* #nosec G501 */
	"encoding/hex"
	"fmt"
	"strings"
)

// GravatarURL Generates Gravatar image URL from email and wanted image height in pixels (results in square-sized image)
func GravatarURL(email string, size int) string {
	cleanedEmail := strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(cleanedEmail)) /* #nosec G401 */
	hex := hex.EncodeToString(hash[:])
	return fmt.Sprintf("http://www.gravatar.com/avatar/%s?d=identicon&s=%d", hex, size)
}
