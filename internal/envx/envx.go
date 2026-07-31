// Package envx reads environment variables with akme-first, nx-compat fallbacks.
package envx

import (
	"os"
	"strings"
)

// First returns the first non-empty environment value among keys.
func First(keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}

// Truthy reports whether First(keys...) is a common truthy flag value.
func Truthy(keys ...string) bool {
	switch strings.ToLower(First(keys...)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
