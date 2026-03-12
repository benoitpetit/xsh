// Package core provides dynamic GraphQL endpoint discovery from X.com
// This implementation extracts operation IDs and feature switches from JS bundles
// similar to the Python version but with Go's concurrency advantages.
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/benoitpetit/xsh/utils"
)

const (
	// CacheTTL is the time-to-live for cached endpoints (24 hours like Python)
	CacheTTL = 24 * time.Hour
	// MaxCacheAge is the maximum age before forced refresh
	MaxCacheAge = 7 * 24 * time.Hour
	// HomepageURL is used to discover JS bundles
	HomepageURL = "https://x.com/elonmusk"
	// BundleCDNBase is the base URL for JS bundles
	BundleCDNBase = "https://abs.twimg.com/responsive-web/client-web"
)

var (
	// Regex patterns for extracting data from HTML/JS
	bundleHrefPattern   = regexp.MustCompile(`href="(https://abs\.twimg\.com/responsive-web/client-web/[^"]+\.js)"`)
	bundleSrcPattern    = regexp.MustCompile(`src="(https://abs\.twimg\.com/responsive-web/client-web/[^"]+\.js)"`)
	chunkMapPattern     = regexp.MustCompile(`"\+(\{[^}]+\})\[e\]\+"a\.js"`)
	operationPattern    = regexp.MustCompile(`queryId:\s*"([A-Za-z0-9_-]+)".*?operationName:\s*"([A-Za-z]+)"`)
	featureSwitchesPattern = regexp.MustCompile(`featureSwitches:\s*(\[[^\]]*\])`)
	
	// Memory cache for in-session performance
	memoryCache     *EndpointCache
	memoryCacheMu   sync.RWMutex
	memoryCacheOnce sync.Once
)

// EndpointCache represents the cached endpoint data
type EndpointCache struct {
	Endpoints   map[string]string   `json:"endpoints"`
	Features    map[string]bool     `json:"features"`
	OpFeatures  map[string][]string `json:"op_features"`
	Timestamp   time.Time           `json:"timestamp"`
	Version     string              `json:"version"`
	Fingerprint string              `json:"fingerprint"` // Hash of X.com response for change detection
}

// GetMemoryCache returns the singleton memory cache
func GetMemoryCache() *EndpointCache {
	memoryCacheOnce.Do(func() {
		memoryCache = &EndpointCache{
			Endpoints:  make(map[string]string),
			Features:   make(map[string]bool),
			OpFeatures: make(map[string][]string),
		}
	})
	return memoryCache
}

// EndpointDiscovery manages dynamic endpoint extraction
type EndpointDiscovery struct {
	client    *http.Client
	cachePath string
	verbose   bool
}

// NewEndpointDiscovery creates a new endpoint discovery instance
func NewEndpointDiscovery(verbose bool) (*EndpointDiscovery, error) {
	cachePath, err := getEndpointCachePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache path: %w", err)
	}

	return &EndpointDiscovery{
		client:    createDiscoveryHTTPClient(),
		cachePath: cachePath,
		verbose:   verbose,
	}, nil
}

// createDiscoveryHTTPClient creates an HTTP client with TLS fingerprinting for endpoint discovery
// Uses uTLS to mimic Chrome browser fingerprint and avoid bot detection
func createDiscoveryHTTPClient() *http.Client {
	// Use uTLS for advanced TLS fingerprinting (like Python's curl_cffi)
	proxy := os.Getenv("X_PROXY")
	if proxy == "" {
		proxy = os.Getenv("TWITTER_PROXY")
	}
	
	client, err := newUTLSHTTPClient(proxy)
	if err != nil {
		// Fallback to standard client if uTLS fails
		return &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		}
	}
	
	return client
}

// DiscoverEndpoints performs full endpoint discovery from X.com
// This is the main entry point equivalent to Python's _fetch_and_extract
func (ed *EndpointDiscovery) DiscoverEndpoints(ctx context.Context) (*EndpointCache, error) {
	if ed.verbose {
		log.Println("[EndpointDiscovery] Starting endpoint discovery from X.com...")
	}

	// Step 1: Fetch homepage to discover JS bundles
	html, fingerprint, err := ed.fetchHomepage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch homepage: %w", err)
	}

	// Step 2: Extract bundle URLs
	bundleURLs := ed.extractBundleURLs(html)
	if len(bundleURLs) == 0 {
		return nil, fmt.Errorf("no JS bundle URLs found in X.com HTML")
	}

	if ed.verbose {
		log.Printf("[EndpointDiscovery] Found %d bundle URLs", len(bundleURLs))
	}

	// Step 3: Download bundles and extract operations concurrently
	endpoints, opFeatures := ed.extractOperationsConcurrent(ctx, bundleURLs)

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no GraphQL operations found in any bundle")
	}

	// Step 4: Extract feature flags from HTML
	features := ed.extractFeaturesFromHTML(html)

	// Build cache
	cache := &EndpointCache{
		Endpoints:   endpoints,
		Features:    features,
		OpFeatures:  opFeatures,
		Timestamp:   time.Now(),
		Version:     "1.0",
		Fingerprint: fingerprint,
	}

	// Save to disk and memory
	if err := ed.SaveCache(cache); err != nil && ed.verbose {
		log.Printf("[EndpointDiscovery] Warning: failed to save cache: %v", err)
	}

	ed.UpdateMemoryCache(cache)

	if ed.verbose {
		log.Printf("[EndpointDiscovery] Discovered %d endpoints, %d features, %d op-feature mappings",
			len(endpoints), len(features), len(opFeatures))
	}

	return cache, nil
}

// fetchHomepage fetches X.com homepage and returns HTML + fingerprint
func (ed *EndpointDiscovery) fetchHomepage(ctx context.Context) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", HomepageURL, nil)
	if err != nil {
		return "", "", err
	}

	// Set headers to mimic browser
	// Note: Don't set Accept-Encoding to avoid compression (Go's http.Client doesn't auto-decompress with uTLS)
	req.Header.Set("User-Agent", GetUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", GetAcceptLanguage())
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("sec-ch-ua", GetSecChUa())
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"`+GetPlatform()+`"`)

	resp, err := ed.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if ed.verbose {
		log.Printf("[EndpointDiscovery] Response status: %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("X.com returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // Max 10MB
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}
	
	if ed.verbose {
		log.Printf("[EndpointDiscovery] Response body size: %d bytes", len(body))
	}

	html := string(body)
	
	// Create fingerprint from first 1KB to detect changes
	fingerprint := utils.HashString(html[:min(len(html), 1024)])

	return html, fingerprint, nil
}

// extractBundleURLs extracts JS bundle URLs from HTML
func (ed *EndpointDiscovery) extractBundleURLs(html string) []string {
	seen := make(map[string]bool)
	var urls []string

	// Extract from href and src attributes
	for _, pattern := range []*regexp.Regexp{bundleHrefPattern, bundleSrcPattern} {
		matches := pattern.FindAllStringSubmatch(html, -1)
		for _, match := range matches {
			if len(match) > 1 && !seen[match[1]] {
				seen[match[1]] = true
				urls = append(urls, match[1])
			}
		}
	}

	// Extract chunk URLs from inline JS mappings
	chunkMatches := chunkMapPattern.FindAllStringSubmatch(html, -1)
	for _, match := range chunkMatches {
		if len(match) > 1 {
			var chunkMap map[string]string
			if err := json.Unmarshal([]byte(match[1]), &chunkMap); err == nil {
				for name, hash := range chunkMap {
					chunkURL := fmt.Sprintf("%s/%s.%sa.js", BundleCDNBase, name, hash)
					if !seen[chunkURL] {
						seen[chunkURL] = true
						urls = append(urls, chunkURL)
					}
				}
			}
		}
	}

	// Prioritize main.js bundles
	return prioritizeBundles(urls)
}

// prioritizeBundles puts main.js bundles first
func prioritizeBundles(urls []string) []string {
	var mainBundles, otherBundles []string
	for _, url := range urls {
		if strings.Contains(url, "/main.") {
			mainBundles = append(mainBundles, url)
		} else {
			otherBundles = append(otherBundles, url)
		}
	}
	return append(mainBundles, otherBundles...)
}

// extractOperationsConcurrent extracts GraphQL operations from bundles concurrently
func (ed *EndpointDiscovery) extractOperationsConcurrent(ctx context.Context, bundleURLs []string) (map[string]string, map[string][]string) {
	endpoints := make(map[string]string)
	opFeatures := make(map[string][]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore to limit concurrent downloads
	semaphore := make(chan struct{}, 5)

	for _, url := range bundleURLs {
		wg.Add(1)
		go func(bundleURL string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			ops, feats, err := ed.extractFromBundle(ctx, bundleURL)
			if err != nil {
				if ed.verbose {
					log.Printf("[EndpointDiscovery] Failed to extract from %s: %v", bundleURL, err)
				}
				return
			}

			mu.Lock()
			for opName, endpoint := range ops {
				if existing, ok := endpoints[opName]; ok && existing != endpoint {
					if ed.verbose {
						log.Printf("[EndpointDiscovery] Duplicate operation %s: %s vs %s", opName, existing, endpoint)
					}
				}
				endpoints[opName] = endpoint
			}
			for opName, features := range feats {
				opFeatures[opName] = features
			}
			mu.Unlock()
		}(url)
	}

	wg.Wait()
	return endpoints, opFeatures
}

// extractFromBundle extracts operations from a single JS bundle
func (ed *EndpointDiscovery) extractFromBundle(ctx context.Context, bundleURL string) (map[string]string, map[string][]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", bundleURL, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("User-Agent", GetUserAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://x.com/")

	resp, err := ed.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	js, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // Max 5MB per bundle
	if err != nil {
		return nil, nil, err
	}

	endpoints, opFeatures := extractOperationsFromJS(string(js))
	return endpoints, opFeatures, nil
}

// extractOperationsFromJS parses GraphQL operations from JS content
func extractOperationsFromJS(js string) (map[string]string, map[string][]string) {
	endpoints := make(map[string]string)
	opFeatures := make(map[string][]string)

	// Split by queryId to avoid cross-operation matching
	blocks := strings.Split(js, `queryId:`)

	for _, block := range blocks[1:] { // Skip first empty split
		// Look for operation name within limited scope
		match := operationPattern.FindStringSubmatch(`queryId:` + block[:min(len(block), 500)])
		if match == nil {
			continue
		}

		queryID := match[1]
		opName := match[2]
		endpoint := fmt.Sprintf("%s/%s", queryID, opName)

		endpoints[opName] = endpoint

		// Extract feature switches from the same block (limited scope)
		fsMatch := featureSwitchesPattern.FindStringSubmatch(block[:min(len(block), 3000)])
		if fsMatch != nil {
			var features []string
			if err := json.Unmarshal([]byte(fsMatch[1]), &features); err == nil {
				opFeatures[opName] = features
			}
		}
	}

	return endpoints, opFeatures
}

// extractFeaturesFromHTML extracts feature flags from __INITIAL_STATE__
func (ed *EndpointDiscovery) extractFeaturesFromHTML(html string) map[string]bool {
	features := make(map[string]bool)

	idx := strings.Index(html, "window.__INITIAL_STATE__=")
	if idx == -1 {
		if ed.verbose {
			log.Println("[EndpointDiscovery] No __INITIAL_STATE__ found in HTML")
		}
		return features
	}

	jsonStart := idx + len("window.__INITIAL_STATE__=")
	jsonStr := extractJSONObject(html, jsonStart)
	if jsonStr == "" {
		return features
	}

	var state map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &state); err != nil {
		if ed.verbose {
			log.Printf("[EndpointDiscovery] Failed to parse __INITIAL_STATE__: %v", err)
		}
		return features
	}

	// Navigate to featureSwitch.defaultConfig or featureSwitch.features
	if featureSwitch, ok := state["featureSwitch"].(map[string]interface{}); ok {
		var featuresObj map[string]interface{}
		if dc, ok := featureSwitch["defaultConfig"].(map[string]interface{}); ok {
			featuresObj = dc
		} else if f, ok := featureSwitch["features"].(map[string]interface{}); ok {
			featuresObj = f
		}

		for key, val := range featuresObj {
			if v, ok := val.(bool); ok {
				features[key] = v
			} else if m, ok := val.(map[string]interface{}); ok {
				if v, ok := m["value"].(bool); ok {
					features[key] = v
				}
			}
		}
	}

	return features
}

// extractJSONObject extracts a complete JSON object using brace counting
func extractJSONObject(text string, start int) string {
	if start >= len(text) || text[start] != '{' {
		return ""
	}

	depth := 0
	inString := false
	escape := false

	for i := start; i < len(text); i++ {
		ch := text[i]
		
		if escape {
			escape = false
			continue
		}
		if ch == '\\' {
			escape = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}

	return ""
}

// GetCachedEndpoints returns cached endpoints, fetching if necessary
func (ed *EndpointDiscovery) GetCachedEndpoints(ctx context.Context) (*EndpointCache, error) {
	// Check memory cache first
	if cache := ed.GetMemoryCache(); cache != nil && cache.IsValid() {
		if ed.verbose {
			log.Println("[EndpointDiscovery] Using memory cache")
		}
		return cache, nil
	}

	// Check disk cache
	if cache, err := ed.LoadCache(); err == nil && cache.IsValid() {
		if ed.verbose {
			log.Println("[EndpointDiscovery] Using disk cache")
		}
		ed.UpdateMemoryCache(cache)
		return cache, nil
	}

	// Fetch fresh
	return ed.DiscoverEndpoints(ctx)
}

// IsValid checks if cache is still valid (not expired)
func (ec *EndpointCache) IsValid() bool {
	if ec == nil || len(ec.Endpoints) == 0 {
		return false
	}
	return time.Since(ec.Timestamp) < CacheTTL
}

// IsStale checks if cache is getting old (over 50% of TTL)
func (ec *EndpointCache) IsStale() bool {
	if ec == nil {
		return true
	}
	return time.Since(ec.Timestamp) > CacheTTL/2
}

// GetEndpoint returns the endpoint for an operation
func (ec *EndpointCache) GetEndpoint(operation string) (string, bool) {
	endpoint, ok := ec.Endpoints[operation]
	return endpoint, ok
}

// GetOpFeatures returns feature switches for an operation
func (ec *EndpointCache) GetOpFeatures(operation string) map[string]bool {
	result := make(map[string]bool)
	
	keys, ok := ec.OpFeatures[operation]
	if !ok {
		return result
	}

	for _, key := range keys {
		if val, ok := ec.Features[key]; ok {
			result[key] = val
		} else {
			result[key] = true // Default to true if not found
		}
	}

	return result
}

// SaveCache saves cache to disk (public for EndpointManager)
func (ed *EndpointDiscovery) SaveCache(cache *EndpointCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(ed.cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(ed.cachePath, data, 0600)
}

// LoadCache loads cache from disk (public for EndpointManager)
func (ed *EndpointDiscovery) LoadCache() (*EndpointCache, error) {
	data, err := os.ReadFile(ed.cachePath)
	if err != nil {
		return nil, err
	}

	var cache EndpointCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// UpdateMemoryCache updates the global memory cache (public for EndpointManager)
func (ed *EndpointDiscovery) UpdateMemoryCache(cache *EndpointCache) {
	memoryCacheMu.Lock()
	defer memoryCacheMu.Unlock()
	
	mc := GetMemoryCache()
	mc.Endpoints = cache.Endpoints
	mc.Features = cache.Features
	mc.OpFeatures = cache.OpFeatures
	mc.Timestamp = cache.Timestamp
	mc.Version = cache.Version
	mc.Fingerprint = cache.Fingerprint
}

// GetMemoryCache retrieves the memory cache (public for EndpointManager)
func (ed *EndpointDiscovery) GetMemoryCache() *EndpointCache {
	memoryCacheMu.RLock()
	defer memoryCacheMu.RUnlock()
	
	mc := GetMemoryCache()
	if mc.Timestamp.IsZero() {
		return nil
	}
	return mc
}

// InvalidateCache clears all caches
func (ed *EndpointDiscovery) InvalidateCache() {
	memoryCacheMu.Lock()
	memoryCache = &EndpointCache{
		Endpoints:  make(map[string]string),
		Features:   make(map[string]bool),
		OpFeatures: make(map[string][]string),
	}
	memoryCacheMu.Unlock()

	os.Remove(ed.cachePath)
}

// getEndpointCachePath returns the path to the endpoint cache file
func getEndpointCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", ConfigDirName, "graphql_ops.json"), nil
}

// GetDynamicGraphQLEndpoints returns endpoints from cache or fetches new ones
func GetDynamicGraphQLEndpoints() map[string]string {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return GraphQLEndpoints // Fallback to static
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cache, err := discovery.GetCachedEndpoints(ctx)
	if err != nil {
		if Verbose {
			log.Printf("[EndpointDiscovery] Failed to get cached endpoints: %v, using static fallback", err)
		}
		return GraphQLEndpoints
	}

	return cache.Endpoints
}

// GetDynamicFeatures returns feature flags from cache
func GetDynamicFeatures() map[string]bool {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return DefaultFeatures
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cache, err := discovery.GetCachedEndpoints(ctx)
	if err != nil {
		return DefaultFeatures
	}

	return cache.Features
}

// GetDynamicOpFeatures returns operation-specific features
func GetDynamicOpFeatures(operation string) []string {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cache, err := discovery.GetCachedEndpoints(ctx)
	if err != nil {
		return nil
	}

	return cache.OpFeatures[operation]
}

// RefreshEndpoints forces a refresh of all endpoints
func RefreshEndpoints() error {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return err
	}

	discovery.InvalidateCache()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	_, err = discovery.DiscoverEndpoints(ctx)
	return err
}

// CheckEndpointHealth checks if the current endpoints are still valid
func CheckEndpointHealth(ctx context.Context, client *XClient) (bool, []string) {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return false, []string{fmt.Sprintf("Failed to create discovery: %v", err)}
	}

	cache, err := discovery.GetCachedEndpoints(ctx)
	if err != nil {
		return false, []string{fmt.Sprintf("Failed to get endpoints: %v", err)}
	}

	var issues []string
	
	// Check if cache is getting stale
	if cache.IsStale() {
		issues = append(issues, "Endpoint cache is getting stale, consider refreshing")
	}

	// Test a few critical endpoints
	criticalEndpoints := []string{"HomeTimeline", "UserByScreenName", "SearchTimeline"}
	for _, op := range criticalEndpoints {
		if _, ok := cache.Endpoints[op]; !ok {
			issues = append(issues, fmt.Sprintf("Critical endpoint %s is missing", op))
		}
	}

	return len(issues) == 0, issues
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
