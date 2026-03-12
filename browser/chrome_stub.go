// +build !windows

package browser

import "fmt"

// decryptChromeCookieWindows decrypts cookies on Windows using DPAPI
// This is a stub for non-Windows platforms. The real implementation is in chrome_windows.go
func decryptChromeCookieWindows(encrypted []byte) (string, error) {
	return "", fmt.Errorf("Windows decryption not available on this platform")
}

// IsChromeAvailableWindows checks if Chrome is available on Windows
// This is a stub for non-Windows platforms.
func IsChromeAvailableWindows() bool {
	return false
}
