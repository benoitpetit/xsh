// Package browser provides automatic cookie extraction from browsers.
package browser

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/benoitpetit/xsh/core"
	"golang.org/x/crypto/pbkdf2"
)

// ChromeCookieExtractor extracts cookies from Chrome-based browsers
type ChromeCookieExtractor struct {
	Name string
	Path string
}

// ChromeCookie represents a cookie from Chrome's cookie database
type ChromeCookie struct {
	HostKey        string `json:"host_key"`
	Name           string `json:"name"`
	Value          string `json:"value"`
	EncryptedValue []byte `json:"encrypted_value"`
	Path           string `json:"path"`
}

// GetDefaultChromePaths returns default Chrome cookie paths by OS
func GetDefaultChromePaths() []string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library/Application Support/Google/Chrome/Default/Cookies"),
			filepath.Join(home, "Library/Application Support/Google/Chrome/Profile 1/Cookies"),
			filepath.Join(home, "Library/Application Support/Google/Chrome/Profile 2/Cookies"),
			filepath.Join(home, "Library/Application Support/BraveSoftware/Brave-Browser/Default/Cookies"),
			filepath.Join(home, "Library/Application Support/Microsoft Edge/Default/Cookies"),
			filepath.Join(home, "Library/Application Support/Chromium/Default/Cookies"),
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		return []string{
			filepath.Join(localAppData, "Google/Chrome/User Data/Default/Network/Cookies"),
			filepath.Join(localAppData, "Google/Chrome/User Data/Default/Cookies"),
			filepath.Join(localAppData, "Google/Chrome/User Data/Profile 1/Cookies"),
			filepath.Join(localAppData, "BraveSoftware/Brave-Browser/User Data/Default/Cookies"),
			filepath.Join(localAppData, "Microsoft/Edge/User Data/Default/Cookies"),
			filepath.Join(localAppData, "Chromium/User Data/Default/Cookies"),
		}
	case "linux":
		paths := []string{
			// Standard Chrome paths
			filepath.Join(home, ".config/google-chrome/Default/Cookies"),
			filepath.Join(home, ".config/google-chrome/Profile 1/Cookies"),
			filepath.Join(home, ".config/google-chrome/Profile 2/Cookies"),
			filepath.Join(home, ".config/chromium/Default/Cookies"),
			filepath.Join(home, ".config/BraveSoftware/Brave-Browser/Default/Cookies"),
			filepath.Join(home, ".config/microsoft-edge/Default/Cookies"),
			// Flatpak Chrome
			filepath.Join(home, ".var/app/com.google.Chrome/config/google-chrome/Default/Cookies"),
			filepath.Join(home, ".var/app/com.google.Chrome/config/google-chrome/Profile 1/Cookies"),
			// Snap Chrome (Ubuntu)
			filepath.Join(home, "snap/chromium/common/chromium/Default/Cookies"),
		}
		return paths
	}
	return nil
}

// ExtractCookies extracts Twitter/X cookies from Chrome
func (c *ChromeCookieExtractor) ExtractCookies() (*core.AuthCredentials, error) {
	return c.ExtractCookiesVerbose(false)
}

// isChromeRunning checks if Chrome is currently running
func isChromeRunning() bool {
	cmd := exec.Command("pgrep", "-x", "chrome")
	err := cmd.Run()
	return err == nil
}

// ExtractCookiesVerbose extracts cookies with optional verbose output
func (c *ChromeCookieExtractor) ExtractCookiesVerbose(verbose bool) (*core.AuthCredentials, error) {
	if verbose {
		fmt.Println("[DEBUG] Starting Chrome cookie extraction...")
	}

	if c.Path == "" {
		paths := GetDefaultChromePaths()
		if verbose {
			fmt.Printf("[DEBUG] Searching for Chrome cookie DB in %d paths...\n", len(paths))
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				c.Path = path
				if verbose {
					fmt.Printf("[DEBUG] Found Chrome cookie DB: %s\n", path)
				}
				break
			} else if verbose {
				fmt.Printf("[DEBUG] Path not found: %s\n", path)
			}
		}
	} else if verbose {
		fmt.Printf("[DEBUG] Using provided path: %s\n", c.Path)
	}

	if c.Path == "" {
		return nil, fmt.Errorf("no Chrome cookie database found")
	}

	// Check if Chrome is running
	if isChromeRunning() {
		if verbose {
			fmt.Println("[DEBUG] WARNING: Chrome appears to be running. Cookie database may be locked.")
		}
		// We'll try anyway, but it might fail
	}

	// Copy database to temp location (Chrome locks it)
	tempFile, err := os.CreateTemp("", "chrome_cookies_*.db")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempDB := tempFile.Name()
	tempFile.Close()

	if err := copyFile(c.Path, tempDB); err != nil {
		os.Remove(tempDB)
		if isChromeRunning() {
			return nil, fmt.Errorf("failed to copy cookie database (Chrome appears to be running - please close Chrome completely and try again): %w", err)
		}
		return nil, fmt.Errorf("failed to copy cookie database: %w", err)
	}
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		return nil, fmt.Errorf("failed to open cookie database: %w", err)
	}
	defer db.Close()

	// Query for Twitter/X cookies
	query := `
		SELECT host_key, name, value, encrypted_value, path 
		FROM cookies 
		WHERE host_key LIKE '%twitter.com' 
		   OR host_key LIKE '%.twitter.com'
		   OR host_key LIKE '%x.com' 
		   OR host_key LIKE '%.x.com'
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cookies: %w", err)
	}
	defer rows.Close()

	cookies := make(map[string]string)
	var authToken, ct0 string
	cookieCount := 0

	for rows.Next() {
		var cookie ChromeCookie
		err := rows.Scan(&cookie.HostKey, &cookie.Name, &cookie.Value, &cookie.EncryptedValue, &cookie.Path)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "[DEBUG] Failed to scan cookie row: %v\n", err)
			}
			continue
		}

		cookieCount++

		// Decrypt value
		var decryptedValue string
		if len(cookie.EncryptedValue) > 0 {
			decryptedValue, err = decryptChromeCookie(cookie.EncryptedValue, verbose)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[DEBUG] Failed to decrypt cookie %s: %v\n", cookie.Name, err)
				}
				// Fallback to plain value
				decryptedValue = cookie.Value
			}
		} else {
			decryptedValue = cookie.Value
		}

		// Skip empty values
		if decryptedValue == "" {
			if verbose {
				fmt.Fprintf(os.Stderr, "[DEBUG] Cookie %s has empty value, skipping\n", cookie.Name)
			}
			continue
		}

		// Sanitize cookie value
		sanitizedValue := core.SanitizeCookieValue(decryptedValue)
		cookies[cookie.Name] = sanitizedValue

		if verbose {
			displayValue := sanitizedValue
			if len(displayValue) > 10 {
				displayValue = displayValue[:10]
			}
			fmt.Fprintf(os.Stderr, "[DEBUG] Found cookie: %s=%s... (host: %s)\n", cookie.Name, displayValue, cookie.HostKey)
		}

		if cookie.Name == "auth_token" {
			authToken = sanitizedValue
		}
		if cookie.Name == "ct0" {
			ct0 = sanitizedValue
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Total cookies found: %d, auth_token: %v, ct0: %v\n", cookieCount, authToken != "", ct0 != "")
	}

	if authToken == "" || ct0 == "" {
		return nil, fmt.Errorf("auth_token or ct0 not found in Chrome cookies. Make sure you're logged into x.com in Chrome")
	}

	return &core.AuthCredentials{
		AuthToken: authToken,
		Ct0:       ct0,
		Cookies:   cookies,
	}, nil
}

// decryptChromeCookie decrypts Chrome's encrypted cookie value
func decryptChromeCookie(encrypted []byte, verbose bool) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}

	switch runtime.GOOS {
	case "darwin":
		return decryptChromeCookieMacOS(encrypted)
	case "windows":
		return decryptChromeCookieWindows(encrypted)
	case "linux":
		return decryptChromeCookieLinux(encrypted, verbose)
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// decryptChromeCookieMacOS decrypts cookies on macOS using Keychain
func decryptChromeCookieMacOS(encrypted []byte) (string, error) {
	if len(encrypted) < 3 {
		return "", fmt.Errorf("invalid encrypted value")
	}

	prefix := string(encrypted[:3])
	if prefix != "v10" && prefix != "v11" {
		return "", fmt.Errorf("unsupported encryption version: %s", prefix)
	}

	encrypted = encrypted[3:]

	key, err := getChromeKeyMacOS()
	if err != nil {
		return "", fmt.Errorf("failed to get Chrome key: %w", err)
	}

	if len(encrypted) < 12 {
		return "", fmt.Errorf("encrypted value too short")
	}

	nonce := encrypted[:12]
	ciphertext := encrypted[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// getChromeKeyMacOS retrieves Chrome's encryption key from macOS Keychain
func getChromeKeyMacOS() ([]byte, error) {
	password := "peanuts"
	salt := []byte("saltysalt")
	key := pbkdf2.Key([]byte(password), salt, 1003, 16, sha1.New)

	// Expand to 32 bytes for AES-256
	fullKey := make([]byte, 32)
	copy(fullKey, key)
	copy(fullKey[16:], key)

	return fullKey, nil
}

// decryptChromeCookieLinux decrypts cookies on Linux using AES-128-CBC
// Chrome on Linux uses AES-128-CBC with PKCS7 padding
func decryptChromeCookieLinux(encrypted []byte, verbose bool) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}

	if len(encrypted) < 3 {
		return "", fmt.Errorf("encrypted value too short")
	}

	prefix := string(encrypted[:3])
	if prefix != "v10" && prefix != "v11" {
		// Not encrypted, return as-is
		return string(encrypted), nil
	}

	ciphertext := encrypted[3:]

	salt := []byte("saltysalt")
	iv := []byte("                ") // 16 spaces

	// Get keys
	v10Key := pbkdf2.Key([]byte("peanuts"), salt, 1, 16, sha1.New)
	
	var keys [][]byte
	keys = append(keys, v10Key)

	// For v11, also try libsecret password and empty key
	if prefix == "v11" {
		libsecretPass, err := getChromePasswordFromLibsecret()
		if err == nil && libsecretPass != "" {
			// Use password as-is (it's base64 encoded string, not decoded bytes)
			v11Key := pbkdf2.Key([]byte(libsecretPass), salt, 1, 16, sha1.New)
			// Try v11 key first
			keys = append([][]byte{v11Key}, keys...)
			if verbose {
				fmt.Println("[DEBUG] Using v11 key from libsecret")
			}
		} else if verbose {
			fmt.Printf("[DEBUG] Could not get password from libsecret: %v\n", err)
			fmt.Println("[DEBUG] You may need to install: python3-secretstorage or libsecret")
		}
		
		// Also try empty key (bug in some Chrome versions)
		emptyKey := pbkdf2.Key([]byte{}, salt, 1, 16, sha1.New)
		keys = append(keys, emptyKey)
	}

	// Try each key with AES-128-CBC
	for i, key := range keys {
		plaintext, err := aesCBCDecrypt(key, iv, ciphertext)
		if err == nil {
			// Chrome 24+ adds a 32-byte integrity check prefix
			// Remove it if present
			if len(plaintext) > 32 {
				// Check if first 32 bytes look like a hash (non-printable)
				// and rest looks like a cookie value
				value := plaintext[32:]
				// Verify it's valid UTF-8
				if isPrintableASCII(value) {
					if verbose {
						fmt.Printf("[DEBUG] Decrypted with key %d, removed 32-byte prefix\n", i)
					}
					return string(value), nil
				}
			}
			if verbose {
				fmt.Printf("[DEBUG] Decrypted with key %d\n", i)
			}
			return plaintext, nil
		}
	}

	if verbose {
		fmt.Printf("[DEBUG] Failed to decrypt with %d different keys\n", len(keys))
	}

	return "", fmt.Errorf("failed to decrypt cookie (Chrome may be using a different encryption method)")
}

// isPrintableASCII checks if string contains only printable ASCII characters
func isPrintableASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Allow printable ASCII (32-126) and common cookie characters
		if c < 32 && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
		if c > 126 {
			return false
		}
	}
	return true
}

// aesCBCDecrypt decrypts using AES-128-CBC with PKCS7 unpadding
func aesCBCDecrypt(key, iv, ciphertext []byte) (string, error) {
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("invalid ciphertext length")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// PKCS7 unpadding
	if len(plaintext) == 0 {
		return "", fmt.Errorf("empty plaintext")
	}

	padLength := int(plaintext[len(plaintext)-1])
	if padLength > aes.BlockSize || padLength == 0 {
		return "", fmt.Errorf("invalid padding")
	}

	// Verify padding
	for i := len(plaintext) - padLength; i < len(plaintext); i++ {
		if int(plaintext[i]) != padLength {
			return "", fmt.Errorf("invalid padding bytes")
		}
	}

	return string(plaintext[:len(plaintext)-padLength]), nil
}

// getChromePasswordFromLibsecret attempts to retrieve Chrome's password from libsecret
func getChromePasswordFromLibsecret() (string, error) {
	// Method 1: Try using Python with secretstorage
	pythonScript := `
import sys
try:
    import secretstorage
    conn = secretstorage.dbus_init()
    collection = secretstorage.get_default_collection(conn)
    for item in collection.get_all_items():
        attrs = item.get_attributes()
        if attrs.get('application') == 'chrome' and 'Safe Storage' in item.get_label():
            secret = item.get_secret()
            if secret:
                if isinstance(secret, bytes):
                    secret = secret.decode('utf-8')
                print(secret, end='')
                sys.exit(0)
    sys.exit(1)
except Exception as e:
    sys.exit(1)
`
	cmd := exec.Command("python3", "-c", pythonScript)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output)), nil
	}

	// Method 2: Try using secret-tool (fallback)
	secretToolPaths := []string{
		"/usr/bin/secret-tool",
		"/usr/local/bin/secret-tool",
		"/bin/secret-tool",
	}
	for _, path := range secretToolPaths {
		if _, err := os.Stat(path); err == nil {
			cmd = exec.Command(path, "lookup", "application", "chrome", "label", "Chrome Safe Storage")
			output, err = cmd.Output()
			if err == nil && len(output) > 0 {
				return strings.TrimSpace(string(output)), nil
			}
			
			// Try alternate label formats
			cmd = exec.Command(path, "lookup", "application", "chrome")
			output, err = cmd.Output()
			if err == nil && len(output) > 0 {
				return strings.TrimSpace(string(output)), nil
			}
		}
	}

	// Method 3: Try with busctl for systems with secret-service but no secret-tool
	busctlScript := `
import sys
try:
    import dbus
    bus = dbus.SessionBus()
    service = bus.get_object('org.freedesktop.secrets', '/org/freedesktop/secrets')
    interface = dbus.Interface(service, 'org.freedesktop.secrets.Service')
    
    # Unlock default collection
    default_path = interface.ReadAlias('default')
    collection = bus.get_object('org.freedesktop.secrets', default_path)
    coll_interface = dbus.Interface(collection, 'org.freedesktop.secrets.Collection')
    
    for item_path in coll_interface.SearchItems({'application': 'chrome'}):
        item = bus.get_object('org.freedesktop.secrets', item_path)
        item_interface = dbus.Interface(item, 'org.freedesktop.DBus.Properties')
        label = item_interface.Get('org.freedesktop.secrets.Item', 'Label')
        if 'Safe Storage' in str(label):
            secret_interface = dbus.Interface(item, 'org.freedesktop.secrets.Item')
            session = interface.OpenSession('plain', dbus.String(''))
            (_, secret_bytes) = secret_interface.GetSecret(session)
            if secret_bytes:
                print(bytes(secret_bytes).decode('utf-8'), end='')
                sys.exit(0)
    sys.exit(1)
except Exception as e:
    sys.exit(1)
`
	cmd = exec.Command("python3", "-c", busctlScript)
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output)), nil
	}

	return "", fmt.Errorf("could not retrieve password from libsecret (tried python3-secretstorage, secret-tool, and dbus)")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0600)
}

// ListAvailableChromeBrowsers returns a list of available Chrome-based browsers
func ListAvailableChromeBrowsers() []string {
	var available []string
	paths := GetDefaultChromePaths()

	browserNames := map[string]string{
		"Chrome":        "chrome",
		"Brave-Browser": "brave",
		"Edge":          "edge",
		"Chromium":      "chromium",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			for key, name := range browserNames {
				if strings.Contains(path, key) {
					available = append(available, name)
					break
				}
			}
		}
	}

	return uniqueStrings(available)
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
