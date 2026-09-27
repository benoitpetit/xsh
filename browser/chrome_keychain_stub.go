//go:build !darwin

package browser

import "fmt"

func getChromeKeyMacOS() ([]byte, error) {
	return nil, fmt.Errorf("macOS Keychain is not available on this platform")
}
