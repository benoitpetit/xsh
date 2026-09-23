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
	"net/url"
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
	// HomepageURL is used to discover JS bundles. The authenticated home page
	// exposes the same responsive-web bundles used by the browser session.
	HomepageURL = "https://x.com/home"
	// BundleCDNBase is the base URL for JS bundles
	BundleCDNBase = "https://abs.twimg.com/responsive-web/client-web"
)

var (
	// Regex patterns for extracting data from HTML/JS
	bundleHrefPattern      = regexp.MustCompile(`href="(https://abs\.twimg\.com/responsive-web/client-web/[^"]+\.js)"`)
	bundleSrcPattern       = regexp.MustCompile(`src="(https://abs\.twimg\.com/responsive-web/client-web/[^"]+\.js)"`)
	chunkMapPattern        = regexp.MustCompile(`"\+(\{[^}]+\})\[e\]\+"a\.js"`)
	operationPattern       = regexp.MustCompile("(?s)queryId\\s*:\\s*[\"'`]([A-Za-z0-9_-]+)[\"'`].{0,500}?operationName\\s*:\\s*[\"'`]([A-Za-z0-9_-]+)[\"'`]")
	featureSwitchesPattern = regexp.MustCompile(`featureSwitches:\s*(\[[^\]]*\])`)
	scriptSrcPattern       = regexp.MustCompile(`(?is)<script[^>]+src\s*=\s*["']([^"']+\.js(?:\?[^"']*)?)["']`)
	jsReferencePattern     = regexp.MustCompile("(?i)[\"'`]((?:https?:)?//[^\"'`\\s]+\\.js(?:\\?[^\"'`\\s]*)?|(?:\\.\\.?/|assets/)[^\"'`\\s]+\\.js(?:\\?[^\"'`\\s]*)?)[\"'`]")
	graphqlURLPattern      = regexp.MustCompile(`(?i)/graphql/([A-Za-z0-9_-]+)/([A-Za-z0-9_-]+)`)

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
	client       *http.Client
	publicClient *http.Client
	cachePath    string
	verbose      bool
}

// NewEndpointDiscovery creates a new endpoint discovery instance
func NewEndpointDiscovery(verbose bool) (*EndpointDiscovery, error) {
	cachePath, err := getEndpointCachePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache path: %w", err)
	}

	return &EndpointDiscovery{
		client:       createDiscoveryHTTPClient(),
		publicClient: createStandardDiscoveryHTTPClient(),
		cachePath:    cachePath,
		verbose:      verbose,
	}, nil
}

func createStandardDiscoveryHTTPClient() *http.Client {
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

// createDiscoveryHTTPClient creates an HTTP client with TLS fingerprinting for endpoint discovery
// Uses uTLS to mimic Chrome browser fingerprint and avoid bot detection
func createDiscoveryHTTPClient() *http.Client {
	// Use uTLS for advanced TLS fingerprinting (like Python's curl_cffi)
	proxy := os.Getenv("X_PROXY")
	if proxy == "" {
		proxy = os.Getenv("TWITTER_PROXY")
	}

	client, err := newUTLSHTTPClient(proxy, BestChromeTarget())
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

	// Modern X.com pages load a small x-web entry bundle which imports the
	// actual application chunks. Expand those references before extracting
	// operations; the old responsive-web page listed all bundles directly.
	bundleURLs = ed.expandBundleURLs(ctx, bundleURLs)

	if ed.verbose {
		log.Printf("[EndpointDiscovery] Found %d bundle URLs", len(bundleURLs))
	}

	// Step 3: Download bundles and extract operations concurrently
	endpoints, opFeatures := ed.extractOperationsConcurrent(ctx, bundleURLs)

	// X.com can serve the new x-web shell to a uTLS fingerprint while a normal
	// browser request receives responsive-web, which still contains persisted
	// GraphQL operation modules. Retry the homepage with a standard HTTP client
	// only when the first extraction produced no operations.
	if len(endpoints) == 0 && ed.publicClient != nil {
		if ed.verbose {
			log.Println("[EndpointDiscovery] No operations in first shell, retrying with standard HTTP client")
		}
		publicHTML, publicFingerprint, publicErr := ed.fetchHomepageWithClient(ctx, ed.publicClient)
		if publicErr == nil {
			publicURLs := ed.expandBundleURLs(ctx, ed.extractBundleURLs(publicHTML))
			publicEndpoints, publicOpFeatures := ed.extractOperationsConcurrent(ctx, publicURLs)
			if len(publicEndpoints) > 0 {
				html = publicHTML
				fingerprint = publicFingerprint
				endpoints = publicEndpoints
				opFeatures = publicOpFeatures
			}
		}
	}

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
	return ed.fetchHomepageWithClient(ctx, ed.client)
}

func (ed *EndpointDiscovery) fetchHomepageWithClient(ctx context.Context, client *http.Client) (string, string, error) {
	if client == nil {
		return "", "", fmt.Errorf("homepage HTTP client unavailable")
	}

	creds := getDiscoveryCredentials()
	req, err := ed.newHomepageRequest(ctx, creds != nil)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("HTTP request failed: %w", err)
	}

	// A stale locally stored session should not prevent discovery: the public
	// home page still exposes the responsive-web bundles used by X.com. Retry
	// once without auth when X rejects the session, while keeping auth for the
	// normal authenticated path.
	if resp.StatusCode == http.StatusUnauthorized && creds != nil {
		resp.Body.Close()
		if ed.verbose {
			log.Println("[EndpointDiscovery] Stored session rejected, retrying public home page")
		}
		req, err = ed.newHomepageRequest(ctx, false)
		if err != nil {
			return "", "", err
		}
		resp, err = client.Do(req)
		if err != nil {
			return "", "", fmt.Errorf("HTTP request failed: %w", err)
		}
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

func (ed *EndpointDiscovery) newHomepageRequest(ctx context.Context, authenticated bool) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, HomepageURL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers to mimic a browser. Don't set Accept-Encoding so Go's client
	// can read the response body consistently with both HTTP clients.
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
	req.Header.Set("sec-ch-ua-arch", `"`+GetArchitecture()+`"`)
	req.Header.Set("sec-ch-ua-bitness", `"64"`)
	req.Header.Set("sec-ch-ua-full-version-list", GetSecChUaFullVersionListForVersion(BestChromeTarget()))
	req.Header.Set("sec-ch-ua-model", `""`)
	req.Header.Set("sec-ch-ua-platform-version", `"`+GetPlatformVersion()+`"`)
	req.Header.Set("x-client-transaction-id", generateTransactionID())
	req.Header.Set("x-twitter-client-language", "en")
	if authenticated {
		applyDiscoveryAuth(req, getDiscoveryCredentials())
	}
	return req, nil
}

// getDiscoveryCredentials loads credentials without triggering browser
// extraction. Endpoint refresh should be safe and predictable when no account
// is configured; the normal CLI authentication flow remains responsible for
// importing browser cookies.
func getDiscoveryCredentials() *AuthCredentials {
	if creds := GetAuthFromEnv(); creds != nil && creds.IsValid() {
		return creds
	}

	creds, err := LoadStoredAuth("")
	if err == nil && creds != nil && creds.IsValid() {
		return creds
	}
	return nil
}

func applyDiscoveryAuth(req *http.Request, creds *AuthCredentials) {
	if req == nil || creds == nil || !creds.IsValid() {
		return
	}

	authToken := creds.GetSanitizedAuthToken()
	ct0 := creds.GetSanitizedCt0()
	if authToken == "" || ct0 == "" {
		return
	}

	req.Header.Set("Authorization", "Bearer "+BearerToken)
	req.Header.Set("Cookie", "auth_token="+authToken+"; ct0="+ct0)
	req.Header.Set("x-csrf-token", ct0)
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-client-language", "en")
}

// extractBundleURLs extracts JS bundle URLs from HTML
func (ed *EndpointDiscovery) extractBundleURLs(html string) []string {
	seen := make(map[string]bool)
	var urls []string
	addURL := func(raw string) {
		raw = strings.ReplaceAll(raw, "&amp;", "&")
		if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
			return
		}
		if !strings.HasSuffix(strings.SplitN(raw, "?", 2)[0], ".js") || seen[raw] {
			return
		}
		seen[raw] = true
		urls = append(urls, raw)
	}

	// Current X.com uses script tags such as x-web/x-web/entry-client-*.js.
	for _, match := range scriptSrcPattern.FindAllStringSubmatch(html, -1) {
		if len(match) > 1 {
			addURL(match[1])
		}
	}

	// Extract from href and src attributes
	for _, pattern := range []*regexp.Regexp{bundleHrefPattern, bundleSrcPattern} {
		matches := pattern.FindAllStringSubmatch(html, -1)
		for _, match := range matches {
			if len(match) > 1 {
				addURL(match[1])
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
					addURL(chunkURL)
				}
			}
		}
	}

	// Prioritize main.js bundles
	return prioritizeBundles(urls)
}

// extractReferencedBundleURLs finds JavaScript imports in an already fetched
// bundle. Vite/Rollup builds used by current X.com commonly emit paths such as
// "assets/chunk-<hash>.js" or "./chunk-<hash>.js".
func extractReferencedBundleURLs(js, baseURL string) []string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	seen := make(map[string]bool)
	var urls []string
	for _, match := range jsReferencePattern.FindAllStringSubmatch(js, -1) {
		if len(match) < 2 {
			continue
		}
		raw := strings.ReplaceAll(match[1], `\u002F`, "/")
		ref, err := url.Parse(raw)
		if err != nil {
			continue
		}
		resolved := base.ResolveReference(ref).String()
		if !strings.HasSuffix(strings.SplitN(resolved, "?", 2)[0], ".js") || seen[resolved] {
			continue
		}
		seen[resolved] = true
		urls = append(urls, resolved)
	}
	return urls
}

// expandBundleURLs follows the first-level imports of entry bundles. The cap
// prevents a malformed page from turning discovery into an unbounded crawl.
func (ed *EndpointDiscovery) expandBundleURLs(ctx context.Context, initial []string) []string {
	const maxBundles = 256
	seen := make(map[string]bool, len(initial))
	urls := make([]string, 0, min(len(initial), maxBundles))
	for _, bundleURL := range initial {
		if !seen[bundleURL] && len(urls) < maxBundles {
			seen[bundleURL] = true
			urls = append(urls, bundleURL)
		}
	}

	for _, bundleURL := range append([]string(nil), urls...) {
		js, err := ed.fetchBundle(ctx, bundleURL)
		if err != nil {
			if ed.verbose {
				log.Printf("[EndpointDiscovery] Failed to inspect imports from %s: %v", bundleURL, err)
			}
			continue
		}
		for _, referenced := range extractReferencedBundleURLs(js, bundleURL) {
			if seen[referenced] || len(urls) >= maxBundles {
				continue
			}
			seen[referenced] = true
			urls = append(urls, referenced)
		}
	}

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
	js, err := ed.fetchBundle(ctx, bundleURL)
	if err != nil {
		return nil, nil, err
	}

	endpoints, opFeatures := extractOperationsFromJS(js)
	return endpoints, opFeatures, nil
}

func (ed *EndpointDiscovery) fetchBundle(ctx context.Context, bundleURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", bundleURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", GetUserAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://x.com/")

	resp, err := ed.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	js, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // Max 5MB per bundle
	if err != nil {
		return "", err
	}

	return string(js), nil
}

// extractOperationsFromJS parses GraphQL operations from JS content
func extractOperationsFromJS(js string) (map[string]string, map[string][]string) {
	endpoints := make(map[string]string)
	opFeatures := make(map[string][]string)

	// Some bundles expose persisted operations as request URLs instead of the
	// older queryId/operationName object shape.
	for _, match := range graphqlURLPattern.FindAllStringSubmatch(js, -1) {
		if len(match) > 2 {
			endpoints[match[2]] = fmt.Sprintf("%s/%s", match[1], match[2])
		}
	}

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
		switch ch {
		case '{':
			depth++
		case '}':
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

// criticalEndpointsForCache returns health-check operations that exist in the
// current X.com bundle. X has renamed/removed the historical HomeTimeline
// operations over time, so a hard-coded check would report a false failure.
func criticalEndpointsForCache(cache *EndpointCache) []string {
	if cache == nil {
		return []string{"HomeTimeline", "UserByScreenName", "SearchTimeline"}
	}

	result := make([]string, 0, 4)
	for _, operation := range []string{
		"HomeTimeline",
		"HomeLatestTimeline",
		"ConnectTabTimeline",
		"GenericTimelineById",
	} {
		if _, ok := cache.Endpoints[operation]; ok {
			result = append(result, operation)
			break
		}
	}

	for _, operation := range []string{"UserByScreenName", "SearchTimeline", "TweetDetail", "UserTweets"} {
		if _, ok := cache.Endpoints[operation]; ok {
			result = append(result, operation)
		}
	}
	return result
}

// RefreshEndpoints forces a refresh of all endpoints
func RefreshEndpoints() error {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return err
	}

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

	// Test a few critical endpoints, adapting to current operation names.
	for _, op := range criticalEndpointsForCache(cache) {
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
