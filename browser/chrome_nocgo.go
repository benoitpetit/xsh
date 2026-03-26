//go:build !cgo
// +build !cgo

package browser

import (
	"fmt"

	"github.com/benoitpetit/xsh/core"
)

// ChromeCookieExtractor is a stub for non-CGO builds
type ChromeCookieExtractor struct {
	Name string
	Path string
}

// ChromeCookie is a stub for non-CGO builds
type ChromeCookie struct {
	HostKey        string `json:"host_key"`
	Name           string `json:"name"`
	Value          string `json:"value"`
	Path           string `json:"path"`
	IsSecure       bool   `json:"is_secure"`
	IsHTTPOnly     bool   `json:"is_http_only"`
	ExpirationDate int64  `json:"expiration_date"`
}

// GetDefaultChromePaths returns empty list in non-CGO builds
func GetDefaultChromePaths() []string {
	return []string{}
}

// ExtractCookies returns error in non-CGO builds
func (c *ChromeCookieExtractor) ExtractCookies() (*core.AuthCredentials, error) {
	return nil, fmt.Errorf("browser cookie extraction requires CGO to be enabled")
}

// ExtractCookiesVerbose returns error in non-CGO builds
func (c *ChromeCookieExtractor) ExtractCookiesVerbose(verbose bool) (*core.AuthCredentials, error) {
	return nil, fmt.Errorf("browser cookie extraction requires CGO to be enabled")
}

// ListAvailableChromeBrowsers returns empty list in non-CGO builds
func ListAvailableChromeBrowsers() []string {
	return []string{}
}

// isChromeRunning always returns false in non-CGO builds
func isChromeRunning() bool {
	return false
}

// decryptChromeCookie returns error in non-CGO builds
func decryptChromeCookie(encrypted []byte, verbose bool) (string, error) {
	return "", fmt.Errorf("cookie decryption requires CGO to be enabled")
}

// decryptChromeCookieMacOS returns error in non-CGO builds
func decryptChromeCookieMacOS(encrypted []byte) (string, error) {
	return "", fmt.Errorf("cookie decryption requires CGO to be enabled")
}

// getChromeKeyMacOS returns error in non-CGO builds
func getChromeKeyMacOS() ([]byte, error) {
	return nil, fmt.Errorf("key retrieval requires CGO to be enabled")
}

// decryptChromeCookieLinux returns error in non-CGO builds
func decryptChromeCookieLinux(encrypted []byte, verbose bool) (string, error) {
	return "", fmt.Errorf("cookie decryption requires CGO to be enabled")
}

// getChromePasswordFromLibsecret returns error in non-CGO builds
func getChromePasswordFromLibsecret() (string, error) {
	return "", fmt.Errorf("password retrieval requires CGO to be enabled")
}

// copyFile returns error in non-CGO builds
func copyFile(src, dst string) error {
	return fmt.Errorf("file copy requires CGO to be enabled")
}

// uniqueStrings returns the input unchanged
func uniqueStrings(s []string) []string {
	return s
}

// isPrintableASCII checks if string is printable ASCII
func isPrintableASCII(s string) bool {
	for _, c := range s {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

// aesCBCDecrypt returns error in non-CGO builds
func aesCBCDecrypt(key, iv, ciphertext []byte) (string, error) {
	return "", fmt.Errorf("decryption requires CGO to be enabled")
}
