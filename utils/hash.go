// Package utils provides helper utilities.
package utils

import (
	"fmt"
	"hash/fnv"
)

// HashString returns a hex-encoded FNV-1a hash of the input string.
// Useful for fingerprinting content to detect changes.
func HashString(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%016x", h.Sum64())
}
