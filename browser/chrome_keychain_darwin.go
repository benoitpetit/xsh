//go:build darwin

package browser

import (
	"crypto/sha1"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// runKeychainCommand is injectable so keychain behavior can be tested without
// invoking the user's login keychain.
var runKeychainCommand = func(args ...string) ([]byte, error) {
	return exec.Command("security", args...).Output()
}

// getChromeKeyMacOS retrieves Chrome's Safe Storage password from Keychain.
func getChromeKeyMacOS() ([]byte, error) {
	output, err := runKeychainCommand("find-generic-password", "-w", "-s", "Chrome Safe Storage")
	if err != nil {
		return nil, fmt.Errorf("macOS Keychain lookup failed: %w", err)
	}
	password := strings.TrimSpace(string(output))
	if password == "" {
		return nil, fmt.Errorf("macOS Keychain returned an empty Chrome Safe Storage password")
	}
	return pbkdf2.Key([]byte(password), []byte("saltysalt"), 1003, 32, sha1.New), nil
}
