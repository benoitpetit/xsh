// Package browser provides unified cookie extraction from all supported browsers
package browser

import (
	"fmt"
	"os"
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

// ExtractFromBrowser extracts cookies from a specific browser
func ExtractFromBrowser(browserName string) (*core.AuthCredentials, error) {
	return ExtractFromBrowserVerbose(browserName, false)
}

// ExtractFromBrowserVerbose extracts cookies with verbose output
func ExtractFromBrowserVerbose(browserName string, verbose bool) (*core.AuthCredentials, error) {
	switch strings.ToLower(browserName) {
	case "chrome", "google-chrome", "google chrome":
		return extractFromChromiumBrowserVerbose("chrome", "Chrome", verbose)
	case "brave", "brave browser":
		return extractFromChromiumBrowserVerbose("brave", "Brave", verbose)
	case "firefox", "mozilla firefox":
		return extractFromFirefoxVerbose(verbose)
	case "edge", "microsoft edge":
		return extractFromChromiumBrowserVerbose("edge", "Edge", verbose)
	case "chromium":
		return extractFromChromiumBrowserVerbose("chromium", "Chromium", verbose)
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
	candidates := DiscoverBrowserCandidates()
	if len(candidates) == 0 {
		return nil, "", fmt.Errorf("no supported browsers found")
	}

	failures := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		creds, err := extractCandidateVerbose(candidate, verbose)
		if err == nil && creds != nil && creds.IsValid() {
			return creds, candidate.Name, nil
		}
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", candidate.Name, err))
		}
	}

	if len(failures) > 0 {
		return nil, "", fmt.Errorf("could not extract valid cookies from any browser (%s)", strings.Join(failures, "; "))
	}
	return nil, "", fmt.Errorf("could not extract valid cookies from any browser")
}

// ListAvailableBrowsers returns a list of available browsers on the system
func ListAvailableBrowsers() []string {
	candidates := DiscoverBrowserCandidates()
	available := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		available = append(available, candidate.Name)
	}
	return available
}

// Browser-specific extractors

func extractFromChromeVerbose(verbose bool) (*core.AuthCredentials, error) {
	return extractFromChromiumBrowserVerbose("chrome", "Chrome", verbose)
}

func extractFromBraveVerbose(verbose bool) (*core.AuthCredentials, error) {
	return extractFromChromiumBrowserVerbose("brave", "Brave", verbose)
}

func extractFromEdgeVerbose(verbose bool) (*core.AuthCredentials, error) {
	return extractFromChromiumBrowserVerbose("edge", "Edge", verbose)
}

func extractFromChromiumVerbose(verbose bool) (*core.AuthCredentials, error) {
	return extractFromChromiumBrowserVerbose("chromium", "Chromium", verbose)
}

func extractFromFirefoxVerbose(verbose bool) (*core.AuthCredentials, error) {
	paths := currentCookiePaths("firefox")
	creds, err := extractFromPaths(paths, func(path string) (*core.AuthCredentials, error) {
		return (&FirefoxCookieExtractor{Path: path}).ExtractCookies()
	})
	if err != nil {
		return nil, fmt.Errorf("Firefox extraction failed: %w", err)
	}
	return creds, nil
}

func extractFromChromiumBrowserVerbose(browserName, displayName string, verbose bool) (*core.AuthCredentials, error) {
	paths := currentCookiePaths(browserName)
	creds, err := extractFromPaths(paths, func(path string) (*core.AuthCredentials, error) {
		return (&ChromeCookieExtractor{Name: displayName, Path: path}).ExtractCookiesVerbose(verbose)
	})
	if err != nil {
		return nil, fmt.Errorf("%s extraction failed: %w", displayName, err)
	}
	return creds, nil
}

func extractCandidateVerbose(candidate BrowserCandidate, verbose bool) (*core.AuthCredentials, error) {
	return extractFromPaths(candidate.CookiePaths, func(path string) (*core.AuthCredentials, error) {
		switch candidate.Name {
		case "firefox":
			return (&FirefoxCookieExtractor{Path: path}).ExtractCookies()
		default:
			return (&ChromeCookieExtractor{Name: candidate.Name, Path: path}).ExtractCookiesVerbose(verbose)
		}
	})
}

func currentCookiePaths(browserName string) []string {
	for _, candidate := range DiscoverBrowserCandidates() {
		if candidate.Name == browserName {
			return candidate.CookiePaths
		}
	}
	return nil
}

func extractFromPaths(paths []string, extract func(path string) (*core.AuthCredentials, error)) (*core.AuthCredentials, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no cookie database found")
	}

	failures := make([]string, 0, len(paths))
	for _, path := range paths {
		creds, err := extract(path)
		if err == nil && creds != nil && creds.IsValid() {
			return creds, nil
		}
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, err))
		}
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("all cookie databases failed (%s)", strings.Join(failures, "; "))
	}
	return nil, fmt.Errorf("no valid credentials found in cookie databases")
}

// IsFirefoxAvailableWindows checks Firefox on Windows
func IsFirefoxAvailableWindows() bool {
	appData := os.Getenv("APPDATA")
	_, err := os.Stat(appData + "/Mozilla/Firefox/Profiles")
	return err == nil
}
