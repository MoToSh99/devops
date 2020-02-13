package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// Exists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func StringArrayToInterfaceArray(a []string) []interface{} {
	b := make([]interface{}, len(a))
	for i, s := range a {
		b[i] = s
	}
	return b
}

func FormatDatetime(timestamp string) string {
	splittedTimestamp := strings.Split(timestamp, ".")
	time := splittedTimestamp[0]
	return time[:16]
}

func GravatarURL(email string, size int) string {
	cleanedEmail := strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(cleanedEmail))
	hex := hex.EncodeToString(hash[:])
	return fmt.Sprintf("http://www.gravatar.com/avatar/%s?d=identicon&s=%d", hex, size)
}
