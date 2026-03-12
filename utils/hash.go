// Package utils provides utility functions for xsh
package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
)

// HashString creates a fast FNV-1a hash of a string (for fingerprints)
func HashString(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return fmt.Sprintf("%016x", h.Sum64())
}

// HashStringSHA256 creates a SHA-256 hash of a string (for secure hashing)
func HashStringSHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// TruncateString truncates a string to max length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
