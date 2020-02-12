package utils

import "os"

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
