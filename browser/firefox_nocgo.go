//go:build !cgo
// +build !cgo

package browser

import (
	"fmt"

	"github.com/benoitpetit/xsh/core"
)

// FirefoxCookieExtractor is a stub for non-CGO builds
type FirefoxCookieExtractor struct {
	Path string
}

// FirefoxCookie is a stub for non-CGO builds
type FirefoxCookie struct {
	Host  string `json:"host"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Path  string `json:"path"`
}

// GetDefaultFirefoxPaths returns empty list in non-CGO builds
func GetDefaultFirefoxPaths() []string {
	return []string{}
}

// ExtractCookies returns error in non-CGO builds
func (f *FirefoxCookieExtractor) ExtractCookies() (*core.AuthCredentials, error) {
	return nil, fmt.Errorf("browser cookie extraction requires CGO to be enabled")
}

// IsFirefoxAvailable always returns false in non-CGO builds
func IsFirefoxAvailable() bool {
	return false
}

// findFirefoxProfiles returns empty list in non-CGO builds
func findFirefoxProfiles(profilesDir string) []string {
	return []string{}
}
