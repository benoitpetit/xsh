// Package browser provides unified cookie extraction from all supported browsers
package browser

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/benoitpetit/xsh/core"
)

func init() {
	// Register the browser extraction function with the core package
	// This enables auto-fallback to browser extraction in GetCredentials()
	core.SetBrowserExtractionFunc(func() (*core.AuthCredentials, string, error) {
		return ExtractFromAllBrowsers()
	})
}

// Extractor is the unified interface for browser cookie extraction
type Extractor interface {
	ExtractCookies() (*core.AuthCredentials, error)
	Name() string
	IsAvailable() bool
}

// ExtractionResult represents the result of a browser extraction attempt
type ExtractionResult struct {
	Browser string
	Creds   *core.AuthCredentials
	Error   error
}

// ExtractFromBrowser extracts cookies from a specific browser
func ExtractFromBrowser(browserName string) (*core.AuthCredentials, error) {
	return ExtractFromBrowserVerbose(browserName, false)
}

// ExtractFromBrowserVerbose extracts cookies with verbose output
func ExtractFromBrowserVerbose(browserName string, verbose bool) (*core.AuthCredentials, error) {
	switch strings.ToLower(browserName) {
	case "chrome", "google-chrome", "google chrome":
		return extractFromChromeVerbose(verbose)
	case "brave", "brave browser":
		return extractFromBraveVerbose(verbose)
	case "firefox", "mozilla firefox":
		return extractFromFirefoxVerbose(verbose)
	case "edge", "microsoft edge":
		return extractFromEdgeVerbose(verbose)
	case "chromium":
		return extractFromChromiumVerbose(verbose)
	default:
		return nil, fmt.Errorf("unsupported browser: %s", browserName)
	}
}

// ExtractFromAllBrowsers tries all available browsers and returns the first valid credentials
func ExtractFromAllBrowsers() (*core.AuthCredentials, string, error) {
	return ExtractFromAllBrowsersVerbose(false)
}

// ExtractFromAllBrowsersVerbose tries all browsers with verbose output
func ExtractFromAllBrowsersVerbose(verbose bool) (*core.AuthCredentials, string, error) {
	availableBrowsers := ListAvailableBrowsers()
	
	if len(availableBrowsers) == 0 {
		return nil, "", fmt.Errorf("no supported browsers found")
	}

	results := make(chan ExtractionResult, len(availableBrowsers))

	// Try all browsers concurrently
	for _, browser := range availableBrowsers {
		go func(b string) {
			creds, err := ExtractFromBrowserVerbose(b, verbose)
			results <- ExtractionResult{
				Browser: b,
				Creds:   creds,
				Error:   err,
			}
		}(browser)
	}

	// Return the first successful result
	for i := 0; i < len(availableBrowsers); i++ {
		result := <-results
		if result.Error == nil && result.Creds != nil && result.Creds.IsValid() {
			return result.Creds, result.Browser, nil
		}
	}

	return nil, "", fmt.Errorf("could not extract valid cookies from any browser")
}

// ListAvailableBrowsers returns a list of available browsers on the system
func ListAvailableBrowsers() []string {
	var available []string

	// Check Chrome-based browsers
	if isChromeAvailable() {
		available = append(available, "chrome")
	}
	if isBraveAvailable() {
		available = append(available, "brave")
	}
	if isEdgeAvailable() {
		available = append(available, "edge")
	}
	if isChromiumAvailable() {
		available = append(available, "chromium")
	}

	// Check Firefox
	if isFirefoxAvailable() {
		available = append(available, "firefox")
	}

	return uniqueStrings(available)
}

// Platform-specific availability checks

func isChromeAvailable() bool {
	switch runtime.GOOS {
	case "windows":
		return IsChromeAvailableWindows()
	case "darwin":
		return isPathExists("/Applications/Google Chrome.app")
	case "linux":
		return isCommandExists("google-chrome") || isCommandExists("chromium")
	}
	return false
}

func isBraveAvailable() bool {
	switch runtime.GOOS {
	case "windows":
		// Check Windows paths
		localAppData := os.Getenv("LOCALAPPDATA")
		return isPathExists(localAppData + "/BraveSoftware/Brave-Browser/User Data/Default")
	case "darwin":
		return isPathExists("/Applications/Brave Browser.app")
	case "linux":
		return isCommandExists("brave") || isPathExists(os.Getenv("HOME")+"/.config/BraveSoftware")
	}
	return false
}

func isEdgeAvailable() bool {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		return isPathExists(localAppData + "/Microsoft/Edge/User Data/Default")
	case "darwin":
		return isPathExists("/Applications/Microsoft Edge.app")
	case "linux":
		return isCommandExists("microsoft-edge")
	}
	return false
}

func isChromiumAvailable() bool {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		return isPathExists(localAppData + "/Chromium/User Data/Default")
	case "darwin":
		return isPathExists("/Applications/Chromium.app")
	case "linux":
		return isCommandExists("chromium") || isCommandExists("chromium-browser")
	}
	return false
}

func isFirefoxAvailable() bool {
	switch runtime.GOOS {
	case "windows":
		return IsFirefoxAvailableWindows()
	case "darwin":
		return isPathExists(os.Getenv("HOME") + "/Library/Application Support/Firefox")
	case "linux":
		return isCommandExists("firefox") || isPathExists(os.Getenv("HOME")+"/.mozilla/firefox")
	}
	return false
}

// Browser-specific extractors

func extractFromChromeVerbose(verbose bool) (*core.AuthCredentials, error) {
	extractor := &ChromeCookieExtractor{Name: "Chrome"}
	creds, err := extractor.ExtractCookiesVerbose(verbose)
	if err != nil {
		return nil, fmt.Errorf("Chrome extraction failed: %w", err)
	}
	return creds, nil
}

func extractFromBraveVerbose(verbose bool) (*core.AuthCredentials, error) {
	// Brave uses same format as Chrome
	var path string
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		path = localAppData + "/BraveSoftware/Brave-Browser/User Data/Default/Network/Cookies"
	case "darwin":
		home, _ := os.UserHomeDir()
		path = home + "/Library/Application Support/BraveSoftware/Brave-Browser/Default/Cookies"
	case "linux":
		home, _ := os.UserHomeDir()
		path = home + "/.config/BraveSoftware/Brave-Browser/Default/Cookies"
	}

	extractor := &ChromeCookieExtractor{Name: "Brave", Path: path}
	creds, err := extractor.ExtractCookiesVerbose(verbose)
	if err != nil {
		return nil, fmt.Errorf("Brave extraction failed: %w", err)
	}
	return creds, nil
}

func extractFromEdgeVerbose(verbose bool) (*core.AuthCredentials, error) {
	var path string
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		path = localAppData + "/Microsoft/Edge/User Data/Default/Network/Cookies"
	case "darwin":
		home, _ := os.UserHomeDir()
		path = home + "/Library/Application Support/Microsoft Edge/Default/Cookies"
	case "linux":
		home, _ := os.UserHomeDir()
		path = home + "/.config/microsoft-edge/Default/Cookies"
	}

	extractor := &ChromeCookieExtractor{Name: "Edge", Path: path}
	creds, err := extractor.ExtractCookiesVerbose(verbose)
	if err != nil {
		return nil, fmt.Errorf("Edge extraction failed: %w", err)
	}
	return creds, nil
}

func extractFromChromiumVerbose(verbose bool) (*core.AuthCredentials, error) {
	var path string
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		path = localAppData + "/Chromium/User Data/Default/Network/Cookies"
	case "darwin":
		home, _ := os.UserHomeDir()
		path = home + "/Library/Application Support/Chromium/Default/Cookies"
	case "linux":
		home, _ := os.UserHomeDir()
		path = home + "/.config/chromium/Default/Cookies"
	}

	extractor := &ChromeCookieExtractor{Name: "Chromium", Path: path}
	creds, err := extractor.ExtractCookiesVerbose(verbose)
	if err != nil {
		return nil, fmt.Errorf("Chromium extraction failed: %w", err)
	}
	return creds, nil
}

func extractFromFirefoxVerbose(verbose bool) (*core.AuthCredentials, error) {
	extractor := &FirefoxCookieExtractor{}
	creds, err := extractor.ExtractCookies()
	if err != nil {
		return nil, fmt.Errorf("Firefox extraction failed: %w", err)
	}
	return creds, nil
}

// Utility functions

func isPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isCommandExists(cmd string) bool {
	// Simple check - look in PATH
	_, err := os.Stat("/usr/bin/" + cmd)
	if err == nil {
		return true
	}
	_, err = os.Stat("/usr/local/bin/" + cmd)
	return err == nil
}

// Platform-specific helpers

// IsFirefoxAvailableWindows checks Firefox on Windows
func IsFirefoxAvailableWindows() bool {
	appData := os.Getenv("APPDATA")
	return isPathExists(appData + "/Mozilla/Firefox/Profiles")
}
