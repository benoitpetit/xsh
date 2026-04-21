// Package core provides transaction ID generation for X.com API requests.
package core

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	transactionCacheFile = "transaction_cache.json"
	transactionCacheTTL  = 3600 // 1 hour
)

// TransactionGenerator generates x-client-transaction-id headers
type TransactionGenerator struct {
	homeHTML       string
	ondemandJS     string
	cachedAt       time.Time
	initialization string
	key            string
	mu             sync.RWMutex
	initialized    bool
}

var (
	globalTG     *TransactionGenerator
	globalTGOnce sync.Once
)

// GetTransactionGenerator returns the global transaction generator instance
func GetTransactionGenerator() *TransactionGenerator {
	globalTGOnce.Do(func() {
		globalTG = &TransactionGenerator{}
	})
	return globalTG
}

// Initialize fetches and caches the necessary data from X.com
func (tg *TransactionGenerator) Initialize(client *http.Client) error {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	if tg.initialized && tg.isCacheValid() {
		return nil
	}

	if tg.loadFromDisk() {
		tg.initialized = true
		return nil
	}

	if err := tg.fetchData(client); err != nil {
		if Verbose {
			logVerbose("Transaction generator fetch failed (using fallback): %v", err)
		}
		tg.generateFallbackData()
	}

	tg.initialized = true
	tg.cachedAt = time.Now()
	tg.saveToDisk()

	return nil
}

// isCacheValid checks if the in-memory cache is still valid
func (tg *TransactionGenerator) isCacheValid() bool {
	if tg.homeHTML == "" || tg.ondemandJS == "" {
		return false
	}
	return time.Since(tg.cachedAt) < transactionCacheTTL*time.Second
}

// getCacheDir returns the directory for cache files
func (tg *TransactionGenerator) getCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "xsh")
}

// loadFromDisk loads cached data from disk
func (tg *TransactionGenerator) loadFromDisk() bool {
	cacheDir := tg.getCacheDir()
	cacheFile := filepath.Join(cacheDir, transactionCacheFile)

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return false
	}

	var cache struct {
		HomeHTML       string    `json:"home_html"`
		OndemandJS     string    `json:"ondemand_js"`
		CachedAt       time.Time `json:"cached_at"`
		Initialization string    `json:"initialization"`
		Key            string    `json:"key"`
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		return false
	}

	if time.Since(cache.CachedAt) > transactionCacheTTL*time.Second {
		return false
	}

	tg.homeHTML = cache.HomeHTML
	tg.ondemandJS = cache.OndemandJS
	tg.cachedAt = cache.CachedAt
	tg.initialization = cache.Initialization
	tg.key = cache.Key

	if Verbose {
		logVerbose("Loaded transaction generator from disk cache")
	}

	return true
}

// saveToDisk saves cached data to disk
func (tg *TransactionGenerator) saveToDisk() error {
	cacheDir := tg.getCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	cache := struct {
		HomeHTML       string    `json:"home_html"`
		OndemandJS     string    `json:"ondemand_js"`
		CachedAt       time.Time `json:"cached_at"`
		Initialization string    `json:"initialization"`
		Key            string    `json:"key"`
	}{
		HomeHTML:       tg.homeHTML,
		OndemandJS:     tg.ondemandJS,
		CachedAt:       tg.cachedAt,
		Initialization: tg.initialization,
		Key:            tg.key,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	cacheFile := filepath.Join(cacheDir, transactionCacheFile)
	return os.WriteFile(cacheFile, data, 0644)
}

// fetchData fetches the homepage and ondemand JS from X.com
func (tg *TransactionGenerator) fetchData(client *http.Client) error {
	req, err := http.NewRequest("GET", BaseURL+"/", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", GetUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", GetAcceptLanguage())
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch homepage: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return err
	}

	tg.homeHTML = string(body)

	ondemandURL := tg.extractOndemandURL(tg.homeHTML)
	if ondemandURL == "" {
		return fmt.Errorf("could not extract ondemand JS URL")
	}

	if Verbose {
		logVerbose("Fetching ondemand JS from: %s", ondemandURL)
	}

	req2, err := http.NewRequest("GET", ondemandURL, nil)
	if err != nil {
		return err
	}

	req2.Header.Set("User-Agent", GetUserAgent())
	req2.Header.Set("Accept", "*/*")
	req2.Header.Set("Referer", BaseURL+"/")

	resp2, err := client.Do(req2)
	if err != nil {
		return fmt.Errorf("failed to fetch ondemand JS: %w", err)
	}
	defer resp2.Body.Close()

	jsBody, err := io.ReadAll(io.LimitReader(resp2.Body, 5*1024*1024))
	if err != nil {
		return err
	}

	tg.ondemandJS = string(jsBody)
	tg.parseInitializationData()

	if Verbose {
		logVerbose("Transaction generator initialized successfully")
	}

	return nil
}

// extractOndemandURL extracts the ondemand JS URL from the homepage HTML
func (tg *TransactionGenerator) extractOndemandURL(html string) string {
	patterns := []string{
		`src="([^"]*ondemand\.s\.[a-f0-9]{20,}\.js)"`,
		`src="([^"]*ondemand[^"]*\.js)"`,
		`"([^"]*ondemand[^"]*\.js)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			url := matches[1]
			if !strings.HasPrefix(url, "http") {
				if strings.HasPrefix(url, "/") {
					url = BaseURL + url
				} else {
					url = BaseURL + "/" + url
				}
			}
			return url
		}
	}

	return ""
}

// parseInitializationData extracts initialization data from the JS
func (tg *TransactionGenerator) parseInitializationData() {
	initPattern := regexp.MustCompile(`["']([a-f0-9]{32,})["']`)
	initMatches := initPattern.FindAllStringSubmatch(tg.ondemandJS, -1)

	if len(initMatches) > 0 {
		tg.initialization = initMatches[0][1]
	}

	keyPattern := regexp.MustCompile(`["']([a-f0-9]{16,})["']`)
	keyMatches := keyPattern.FindAllStringSubmatch(tg.ondemandJS, -1)

	if len(keyMatches) > 1 {
		tg.key = keyMatches[1][1]
	} else if len(keyMatches) > 0 {
		tg.key = keyMatches[0][1]
	}

	if tg.initialization == "" {
		tg.initialization = generateRandomHex(32)
	}
	if tg.key == "" {
		tg.key = generateRandomHex(16)
	}
}

// generateFallbackData generates fallback data when fetch fails
func (tg *TransactionGenerator) generateFallbackData() {
	tg.initialization = generateRandomHex(32)
	tg.key = generateRandomHex(16)
	tg.homeHTML = "<html></html>"
	tg.ondemandJS = "// fallback"
}

// Generate creates a transaction ID for the given method and path
func (tg *TransactionGenerator) Generate(method, path string) string {
	tg.mu.RLock()
	defer tg.mu.RUnlock()

	if tg.initialization == "" {
		return generateSimpleTransactionID()
	}

	timestamp := time.Now().UnixNano()
	hashInput := fmt.Sprintf("%s|%s|%s|%s|%d", method, path, tg.initialization, tg.key, timestamp)
	hash := sha256.Sum256([]byte(hashInput))

	combined := tg.initialization + tg.key + fmt.Sprintf("%x", hash[:16])
	encoded := base64.StdEncoding.EncodeToString([]byte(combined))

	for len(encoded) < 90 {
		encoded += "A"
	}

	randomSuffix := generateRandomString(8)
	encoded = encoded[:len(encoded)-8] + randomSuffix

	return encoded
}

// generateSimpleTransactionID generates a simple transaction ID when proper initialization fails
func generateSimpleTransactionID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	b := make([]byte, 92)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// generateRandomHex generates a random hex string of the specified length
func generateRandomHex(length int) string {
	const hexChars = "abcdef0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = hexChars[rand.Intn(len(hexChars))]
	}
	return string(b)
}

// generateRandomString generates a random string from charset
func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// NeedsTransactionID returns true for operations that require x-client-transaction-id
func NeedsTransactionID(operation string) bool {
	ops := map[string]bool{
		"SearchTimeline":           true,
		"HomeTimeline":             true,
		"HomeLatestTimeline":       true,
		"UserTweets":               true,
		"UserTweetsAndReplies":     true,
		"TweetDetail":              true,
		"UserByScreenName":         true,
		"Followers":                true,
		"Following":                true,
		"BookmarkSearchTimeline":   true,
		"Likes":                    true,
		"ListLatestTweetsTimeline": true,
		"ListMembers":              true,
		"Trends":                   true,
		"ExplorePage":              true,
	}
	return ops[operation]
}

// GenerateTransactionIDForRequest generates a transaction ID for a specific request
func GenerateTransactionIDForRequest(method, path, operation string) string {
	if IsWriteOperation(operation) {
		return ""
	}
	if !NeedsTransactionID(operation) && operation != "" {
		return ""
	}

	tg := GetTransactionGenerator()

	if !tg.initialized {
		client := &http.Client{
			Timeout: 15 * time.Second,
		}
		_ = tg.Initialize(client)
	}

	return tg.Generate(method, path)
}

// InitializeTransactionGenerator initializes the global transaction generator
func InitializeTransactionGenerator(client *http.Client) error {
	tg := GetTransactionGenerator()
	return tg.Initialize(client)
}

// math/rand is auto-seeded since Go 1.20; no init needed.
