package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// AuthCredentials stores authentication credentials
type AuthCredentials struct {
	AuthToken   string            `json:"auth_token"`
	Ct0         string            `json:"ct0"`
	Cookies     map[string]string `json:"cookies,omitempty"`
	AccountName string            `json:"account_name,omitempty"`
}

// IsValid checks if credentials look valid (non-empty)
func (a *AuthCredentials) IsValid() bool {
	return a != nil && a.AuthToken != "" && a.Ct0 != ""
}

// SanitizeCookieValue removes invalid characters from cookie values according to RFC 6265.
// Go's net/http package strictly enforces valid cookie characters and rejects:
// - Double quotes (")
// - Control characters
// - Commas, semicolons, backslashes in certain contexts
func SanitizeCookieValue(value string) string {
	// Trim whitespace
	value = strings.TrimSpace(value)

	// Remove surrounding quotes if present
	value = strings.Trim(value, `"`)
	// Removing a quote can expose whitespace that was outside the original
	// value. Trim again so sanitization is idempotent.
	value = strings.TrimSpace(value)

	// Filter out invalid characters
	// Valid cookie value characters per RFC 6265: printable ASCII except DQUOTE, comma, semicolon, backslash
	var result strings.Builder
	for _, r := range value {
		// Allow printable ASCII characters (32-126) except double quote (34), comma (44), semicolon (59), backslash (92)
		if r >= 32 && r <= 126 && r != 34 && r != 44 && r != 59 && r != 92 {
			result.WriteRune(r)
		}
	}

	return strings.TrimSpace(result.String())
}

// SanitizeCookies sanitizes all cookie values in a map
func SanitizeCookies(cookies map[string]string) map[string]string {
	if cookies == nil {
		return nil
	}
	sanitized := make(map[string]string, len(cookies))
	for k, v := range cookies {
		sanitized[k] = SanitizeCookieValue(v)
	}
	return sanitized
}

// GetSanitizedAuthToken returns the sanitized auth_token
func (a *AuthCredentials) GetSanitizedAuthToken() string {
	return SanitizeCookieValue(a.AuthToken)
}

// GetSanitizedCt0 returns the sanitized ct0
func (a *AuthCredentials) GetSanitizedCt0() string {
	return SanitizeCookieValue(a.Ct0)
}

// GetSanitizedCookies returns all cookies with sanitized values
func (a *AuthCredentials) GetSanitizedCookies() map[string]string {
	cookies := SanitizeCookies(a.Cookies)
	if cookies == nil {
		cookies = make(map[string]string)
	}
	cookies["auth_token"] = a.GetSanitizedAuthToken()
	cookies["ct0"] = a.GetSanitizedCt0()
	return cookies
}

// AuthError is raised when authentication fails
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// AuthData represents the stored auth file structure
type AuthData struct {
	Default  string                      `json:"default"`
	Accounts map[string]*AuthCredentials `json:"accounts"`
}

// GetAuthFile returns the path to auth credentials file
func GetAuthFile() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, AuthFileName), nil
}

// LoadStoredAuth loads credentials from stored auth file
func LoadStoredAuth(account string) (*AuthCredentials, error) {
	store, err := defaultAuthStore()
	if err != nil {
		return nil, err
	}
	return store.Load(account)
}

// SaveAuth saves credentials to auth file
func SaveAuth(creds *AuthCredentials, account string) error {
	store, err := defaultAuthStore()
	if err != nil {
		return err
	}
	return store.Save(creds, account)
}

// ImportCookiesFromFile imports cookies from a Cookie Editor JSON export file
func ImportCookiesFromFile(filePath string) (*AuthCredentials, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Cookie Editor exports a list of cookie objects
	var cookies []struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Domain string `json:"domain"`
	}

	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, err
	}

	cookieMap := make(map[string]string)
	for _, c := range cookies {
		domain := strings.ToLower(strings.TrimSpace(c.Domain))
		if isAllowedCookieDomain(domain) && isValidCookieImportValue(c.Value) {
			cookieMap[c.Name] = c.Value
		}
	}

	authToken := cookieMap["auth_token"]
	ct0 := cookieMap["ct0"]

	if authToken == "" || ct0 == "" {
		return nil, fmt.Errorf("auth_token or ct0 not found in cookies")
	}

	return &AuthCredentials{
		AuthToken: authToken,
		Ct0:       ct0,
		Cookies:   cookieMap,
	}, nil
}

func isAllowedCookieDomain(domain string) bool {
	return domain == "x.com" || domain == "twitter.com" ||
		strings.HasSuffix(domain, ".x.com") || strings.HasSuffix(domain, ".twitter.com")
}

func isValidCookieImportValue(value string) bool {
	return value != "" && SanitizeCookieValue(value) == value
}

// ListAccounts lists all stored account names
func ListAccounts() ([]string, error) {
	store, err := defaultAuthStore()
	if err != nil {
		return nil, err
	}
	return store.List()
}

// SetDefaultAccount sets the default account
func SetDefaultAccount(account string) error {
	store, err := defaultAuthStore()
	if err != nil {
		return err
	}
	return store.SetDefault(account)
}

// RemoveAuth removes an account from storage
func RemoveAuth(account string) error {
	store, err := defaultAuthStore()
	if err != nil {
		return err
	}
	if _, loadErr := store.Load(account); loadErr != nil {
		return loadErr
	}
	return store.Remove(account)
}

// GetAuthFromEnv gets credentials from environment variables
func GetAuthFromEnv() *AuthCredentials {
	authToken := os.Getenv("X_AUTH_TOKEN")
	if authToken == "" {
		authToken = os.Getenv("TWITTER_AUTH_TOKEN")
	}

	ct0 := os.Getenv("X_CT0")
	if ct0 == "" {
		ct0 = os.Getenv("TWITTER_CT0")
	}

	if authToken != "" && ct0 != "" {
		return &AuthCredentials{
			AuthToken: authToken,
			Ct0:       ct0,
		}
	}
	return nil
}

// GetCredentials gets credentials using priority: env vars > stored > browser > error
func GetCredentials(account string) (*AuthCredentials, error) {
	cfg, _ := LoadConfig()
	selectedAccount := account
	if selectedAccount == "" && cfg != nil {
		selectedAccount = cfg.DefaultAccount
	}

	// Explicit and configured accounts take precedence over environment and
	// browser fallbacks. An empty selection may still use environment auth.
	if selectedAccount != "" {
		creds, err := LoadStoredAuth(selectedAccount)
		if err != nil {
			return nil, err
		}
		if creds != nil && creds.IsValid() {
			return creds, nil
		}
	} else if creds := GetAuthFromEnv(); creds != nil && creds.IsValid() {
		return creds, nil
	}

	// Stored default is used after environment auth when no account was
	// explicitly/configured, preserving the historical fallback behavior.
	if selectedAccount == "" {
		creds, err := LoadStoredAuth("")
		if err != nil {
			return nil, err
		}
		if creds != nil && creds.IsValid() {
			return creds, nil
		}
	}

	// 3. Browser extraction (auto-fallback like Python)
	// Try to extract from available browsers automatically
	creds, browserName, err := tryBrowserExtraction()
	if err == nil && creds != nil && creds.IsValid() {
		// Save extracted credentials under the requested/configured account.
		saveAccount := selectedAccount
		if saveAccount == "" {
			saveAccount = "default"
		}
		creds.AccountName = saveAccount
		if saveErr := SaveAuth(creds, saveAccount); saveErr == nil {
			return creds, nil
		}
		// Even if save fails, return the creds for this session
		return creds, nil
	}

	// 4. No credentials found
	return nil, &AuthError{
		Message: fmt.Sprintf(`No Twitter/X credentials found. Options:
  1. Set X_AUTH_TOKEN and X_CT0 environment variables
  2. Run 'xsh auth login' to extract from browser
  3. Run 'xsh auth import <file>' to import from Cookie Editor
  4. Run 'xsh auth set' to manually enter credentials

Note: Tried browsers: %v`, browserName),
	}
}

// tryBrowserExtraction attempts to extract credentials from browsers
// This is a placeholder - the actual implementation imports the browser package
// and is set via init() to avoid circular imports
var tryBrowserExtraction = func() (*AuthCredentials, string, error) {
	return nil, "", fmt.Errorf("browser extraction not available")
}

// SetBrowserExtractionFunc allows the browser package to register its extraction function
func SetBrowserExtractionFunc(fn func() (*AuthCredentials, string, error)) {
	tryBrowserExtraction = fn
}

// GetBrowserCookiePaths returns possible browser cookie database paths based on OS
func GetBrowserCookiePaths() map[string][]string {
	home, _ := os.UserHomeDir()
	paths := make(map[string][]string)

	switch runtime.GOOS {
	case "darwin":
		paths["chrome"] = []string{
			home + "/Library/Application Support/Google/Chrome/Default/Cookies",
			home + "/Library/Application Support/Google/Chrome/Profile */Cookies",
		}
		paths["firefox"] = []string{
			home + "/Library/Application Support/Firefox/Profiles/*/cookies.sqlite",
		}
		paths["edge"] = []string{
			home + "/Library/Application Support/Microsoft Edge/Default/Cookies",
		}
		paths["brave"] = []string{
			home + "/Library/Application Support/BraveSoftware/Brave-Browser/Default/Cookies",
		}
	case "linux":
		paths["chrome"] = []string{
			home + "/.config/google-chrome/Default/Cookies",
			home + "/.config/chromium/Default/Cookies",
		}
		paths["firefox"] = []string{
			home + "/.mozilla/firefox/*/cookies.sqlite",
		}
		paths["edge"] = []string{
			home + "/.config/microsoft-edge/Default/Cookies",
		}
		paths["brave"] = []string{
			home + "/.config/BraveSoftware/Brave-Browser/Default/Cookies",
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		paths["chrome"] = []string{
			localAppData + "/Google/Chrome/User Data/Default/Network/Cookies",
			localAppData + "/Google/Chrome/User Data/Default/Cookies",
		}
		paths["firefox"] = []string{
			home + "/AppData/Roaming/Mozilla/Firefox/Profiles/*/cookies.sqlite",
		}
		paths["edge"] = []string{
			localAppData + "/Microsoft/Edge/User Data/Default/Network/Cookies",
			localAppData + "/Microsoft/Edge/User Data/Default/Cookies",
		}
		paths["brave"] = []string{
			localAppData + "/BraveSoftware/Brave-Browser/User Data/Default/Network/Cookies",
		}
	}

	return paths
}

// Note: Browser cookie extraction requires CGO and SQLite bindings
// For pure Go, we'll provide manual cookie import only
// Users can use browser extensions like "Cookie-Editor" to export cookies

// ExtractCookiesManual prompts user for manual cookie entry
func ExtractCookiesManual() (*AuthCredentials, error) {
	return nil, &AuthError{
		Message: "Browser cookie extraction requires manual import. Use 'xsh auth import' with a Cookie Editor export, or 'xsh auth set' to enter tokens manually.",
	}
}
