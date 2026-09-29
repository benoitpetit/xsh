// Package core provides dynamic GraphQL endpoint discovery from X.com
// This implementation extracts operation IDs and feature switches from JS bundles
// similar to the Python version but with Go's concurrency advantages.
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	bundleHrefPattern             = regexp.MustCompile(`href="(https://abs\.twimg\.com/responsive-web/client-web/[^"]+\.js)"`)
	bundleSrcPattern              = regexp.MustCompile(`src="(https://abs\.twimg\.com/responsive-web/client-web/[^"]+\.js)"`)
	chunkMapPattern               = regexp.MustCompile(`"\+(\{[^}]+\})\[e\]\+"a\.js"`)
	operationPattern              = regexp.MustCompile("(?s)queryId\\s*:\\s*[\"'`]([A-Za-z0-9_-]+)[\"'`].{0,500}?operationName\\s*:\\s*[\"'`]([A-Za-z0-9_-]+)[\"'`]")
	operationRecordPattern        = regexp.MustCompile("(?s)(?:[\"'`]?queryId[\"'`]?\\s*[:=]\\s*[\"'`]([A-Za-z0-9_-]+)[\"'`]).{0,1000}?(?:[\"'`]?operationName[\"'`]?\\s*[:=]\\s*[\"'`]([A-Za-z0-9_-]+)[\"'`])")
	reverseOperationRecordPattern = regexp.MustCompile("(?s)(?:[\"'`]?operationName[\"'`]?\\s*[:=]\\s*[\"'`]([A-Za-z0-9_-]+)[\"'`]).{0,1000}?(?:[\"'`]?queryId[\"'`]?\\s*[:=]\\s*[\"'`]([A-Za-z0-9_-]+)[\"'`])")
	queryIDAssignmentPattern      = regexp.MustCompile("(?i)(?:const|let|var)\\s+([A-Za-z_$][A-Za-z0-9_$]*)\\s*=\\s*[\"'`]([A-Za-z0-9_-]+)[\"'`]")
	generatedGraphQLURLPattern    = regexp.MustCompile("(?i)[\"'`](?:/i/api)?/graphql/[\"'`]\\s*\\+\\s*([A-Za-z_$][A-Za-z0-9_$]*)\\s*\\+\\s*[\"'`]/([A-Za-z0-9_%-]+)[\"'`]")
	featureSwitchesPattern        = regexp.MustCompile(`featureSwitches:\s*(\[[^\]]*\])`)
	scriptSrcPattern              = regexp.MustCompile(`(?is)<script[^>]+src\s*=\s*["']([^"']+\.js(?:\?[^"']*)?)["']`)
	jsReferencePattern            = regexp.MustCompile("(?i)[\"'`]((?:https?:)?//[^\"'`\\s]+\\.js(?:\\?[^\"'`\\s]*)?|(?:\\.\\.?/|assets/)[^\"'`\\s]+\\.js(?:\\?[^\"'`\\s]*)?)[\"'`]")
	graphqlURLPattern             = regexp.MustCompile(`(?i)(?:/i/api)?/graphql/([A-Za-z0-9_%-]+)(?:/|%2f)([A-Za-z0-9_%-]+)`)

	// Memory cache for in-session performance
	memoryCache   *EndpointCache
	memoryCacheMu sync.RWMutex
)

// EndpointCache represents the cached endpoint data
type EndpointCache struct {
	Endpoints      map[string]string   `json:"endpoints"`
	Quarantined    map[string]string   `json:"quarantined,omitempty"`
	QuarantinedIDs map[string]string   `json:"quarantined_ids,omitempty"`
	Features       map[string]bool     `json:"features"`
	OpFeatures     map[string][]string `json:"op_features"`
	Timestamp      time.Time           `json:"timestamp"`
	Version        string              `json:"version"`
	Fingerprint    string              `json:"fingerprint"` // Hash of X.com response for change detection
}

// GetMemoryCache returns the singleton memory cache
func GetMemoryCache() *EndpointCache {
	memoryCacheMu.RLock()
	defer memoryCacheMu.RUnlock()
	return cloneEndpointCache(memoryCache)
}

// EndpointDiscovery manages dynamic endpoint extraction
type EndpointDiscovery struct {
	client       *http.Client
	publicClient *http.Client
	cachePath    string
	verbose      bool
	credentials  *AuthCredentials
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

// NewEndpointDiscoveryForAccount selects a stored account without importing
// browser cookies. An empty account uses the configured default.
func NewEndpointDiscoveryForAccount(verbose bool, account string) (*EndpointDiscovery, error) {
	ed, err := NewEndpointDiscovery(verbose)
	if err != nil || account == "" {
		return ed, err
	}
	creds, err := LoadStoredAuth(account)
	if err != nil {
		return nil, err
	}
	if creds == nil || !creds.IsValid() {
		return nil, fmt.Errorf("no stored credentials for account %q", account)
	}
	ed.credentials = creds
	return ed, nil
}

func (ed *EndpointDiscovery) discoveryCredentials() *AuthCredentials {
	if ed != nil && ed.credentials != nil {
		return ed.credentials
	}
	return getDiscoveryCredentials()
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
	proxy := ""
	if proxyURL, err := ResolveProxy("", os.Getenv); err == nil && proxyURL != nil {
		proxy = proxyURL.String()
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
	return ed.discoverEndpoints(ctx, true)
}

// PreviewEndpoints extracts current operations without changing the active cache.
func (ed *EndpointDiscovery) PreviewEndpoints(ctx context.Context) (*EndpointCache, error) {
	return ed.discoverEndpoints(ctx, false)
}

func (ed *EndpointDiscovery) discoverEndpoints(ctx context.Context, publish bool) (*EndpointCache, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ed.verbose {
		log.Println("[EndpointDiscovery] Starting endpoint discovery from X.com...")
	}
	if ed.discoveryCredentials() == nil {
		return nil, fmt.Errorf("authenticated X credentials required for endpoint discovery")
	}

	// Step 1: Fetch homepage to discover JS bundles
	html, fingerprint, err := ed.fetchHomepage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch homepage: %w", err)
	}
	if isLoggedOutShell(html) {
		return nil, fmt.Errorf("X returned a logged-out shell; authenticated endpoint discovery is unavailable")
	}

	// Step 2: Extract bundle URLs
	bundleURLs := ed.extractBundleURLs(html)
	if len(bundleURLs) == 0 {
		return nil, fmt.Errorf("no JS bundle URLs found in X.com HTML")
	}

	// Modern X.com pages load a small x-web entry bundle which imports the
	// actual application chunks. Fetch each bundle once, following imports and
	// extracting operations from the same response.
	bundleURLs, endpoints, opFeatures := ed.discoverBundleOperations(ctx, bundleURLs)
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if ed.verbose {
		log.Printf("[EndpointDiscovery] Found %d bundle URLs", len(bundleURLs))
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no GraphQL operations found in any bundle")
	}

	// Step 4: Extract feature flags from HTML
	features := ed.extractFeaturesFromHTML(html)

	// Build cache
	cache := &EndpointCache{
		Endpoints:      endpoints,
		Quarantined:    make(map[string]string),
		QuarantinedIDs: make(map[string]string),
		Features:       features,
		OpFeatures:     opFeatures,
		Timestamp:      time.Now(),
		Version:        "1.0",
		Fingerprint:    fingerprint,
	}

	if publish {
		if err := ed.SaveCache(cache); err != nil {
			return nil, fmt.Errorf("failed to persist discovered endpoints: %w", err)
		}
		ed.UpdateMemoryCache(cache)
	}

	if ed.verbose {
		log.Printf("[EndpointDiscovery] Discovered %d endpoints, %d features, %d op-feature mappings",
			len(endpoints), len(features), len(opFeatures))
	}

	return cache, nil
}

// fetchHomepage fetches X.com homepage and returns HTML + fingerprint
func (ed *EndpointDiscovery) fetchHomepage(ctx context.Context) (string, string, error) {
	var lastErr error
	for _, page := range []string{HomepageURL, BaseURL + "/"} {
		html, fingerprint, err := ed.fetchHomepageWithClientAt(ctx, ed.client, page)
		if err == nil && !isLoggedOutShell(html) {
			return html, fingerprint, nil
		}
		if err == nil {
			err = fmt.Errorf("X returned a logged-out shell")
		}
		lastErr = err
		if ctx.Err() != nil {
			return "", "", ctx.Err()
		}
	}
	return "", "", lastErr
}

func (ed *EndpointDiscovery) fetchHomepageWithClient(ctx context.Context, client *http.Client) (string, string, error) {
	return ed.fetchHomepageWithClientAt(ctx, client, HomepageURL)
}

func (ed *EndpointDiscovery) fetchHomepageWithClientAt(ctx context.Context, client *http.Client, page string) (string, string, error) {
	if client == nil {
		return "", "", fmt.Errorf("homepage HTTP client unavailable")
	}

	creds := ed.discoveryCredentials()
	req, err := ed.newHomepageRequestAt(ctx, creds != nil, page)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Do(req)
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

	body, err := ReadLimitedBody(resp.Body, 10*1024*1024)
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
	return ed.newHomepageRequestAt(ctx, authenticated, HomepageURL)
}

func (ed *EndpointDiscovery) newHomepageRequestAt(ctx context.Context, authenticated bool, page string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, page, nil)
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
	if authenticated {
		applyDiscoveryBrowserCookies(req, ed.discoveryCredentials())
	}
	return req, nil
}

// getDiscoveryCredentials loads credentials without triggering browser
// extraction. Endpoint refresh should be safe and predictable when no account
// is configured; the normal CLI authentication flow remains responsible for
// importing browser cookies.
func getDiscoveryCredentials() *AuthCredentials {
	// Keep discovery aligned with the account selected by normal CLI requests.
	// Do not invoke GetCredentials here: its browser import has side effects.
	if cfg, err := LoadConfig(); err == nil && cfg != nil && cfg.DefaultAccount != "" {
		if creds, err := LoadStoredAuth(cfg.DefaultAccount); err == nil && creds != nil && creds.IsValid() {
			return creds
		}
	}
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
	applyDiscoveryBrowserCookies(req, creds)
	req.Header.Set("x-csrf-token", ct0)
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-client-language", "en")
}

func applyDiscoveryBrowserCookies(req *http.Request, creds *AuthCredentials) {
	if req == nil || creds == nil || !creds.IsValid() {
		return
	}
	authToken := creds.GetSanitizedAuthToken()
	ct0 := creds.GetSanitizedCt0()
	if authToken == "" || ct0 == "" {
		return
	}
	cookies := creds.GetSanitizedCookies()
	cookieNames := make([]string, 0, len(cookies))
	for name := range cookies {
		cookieNames = append(cookieNames, name)
	}
	sort.Strings(cookieNames)
	cookieParts := make([]string, 0, len(cookieNames))
	for _, name := range cookieNames {
		if value := cookies[name]; value != "" {
			cookieParts = append(cookieParts, name+"="+value)
		}
	}
	if len(cookieParts) == 0 {
		cookieParts = []string{"auth_token=" + authToken, "ct0=" + ct0}
	}
	req.Header.Set("Cookie", strings.Join(cookieParts, "; "))
}

func isLoggedOutShell(html string) bool {
	lower := strings.ToLower(html)
	for _, marker := range []string{
		"entry-client-logged-out-",
		"mode=login",
		"/onboarding/",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
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

const (
	maxDiscoveryBundles = 256
	bundleWorkerCount   = 5
)

type bundleDiscoveryResult struct {
	operations map[string]string
	features   map[string][]string
	imports    []string
	err        error
}

// expandBundleURLs follows the imports reachable from the page bundles.
// Discovery itself uses discoverBundleOperations to avoid downloading them
// again after expansion.
func (ed *EndpointDiscovery) expandBundleURLs(ctx context.Context, initial []string) []string {
	urls, _, _ := ed.discoverBundleOperations(ctx, initial)
	return urls
}

// discoverBundleOperations fetches each reachable bundle once, using a bounded
// worker pool for each breadth-first import wave. Workers discard JS bodies as
// soon as operation and import metadata have been extracted, so memory remains
// bounded by the worker count rather than the total bundle count.
func (ed *EndpointDiscovery) discoverBundleOperations(ctx context.Context, initial []string) ([]string, map[string]string, map[string][]string) {
	if ctx == nil {
		ctx = context.Background()
	}
	urls := make([]string, 0, min(len(initial), maxDiscoveryBundles))
	seen := make(map[string]struct{}, len(initial))
	for _, bundleURL := range initial {
		if _, exists := seen[bundleURL]; exists || len(urls) >= maxDiscoveryBundles {
			continue
		}
		seen[bundleURL] = struct{}{}
		urls = append(urls, bundleURL)
	}

	results := make(map[string]bundleDiscoveryResult, len(urls))
	for start := 0; start < len(urls) && start < maxDiscoveryBundles; {
		if err := ctx.Err(); err != nil {
			break
		}
		end := len(urls)
		if end > maxDiscoveryBundles {
			end = maxDiscoveryBundles
		}
		wave := append([]string(nil), urls[start:end]...)
		waveResults := ed.processBundleWave(ctx, wave)
		for _, bundleURL := range wave {
			result, ok := waveResults[bundleURL]
			if !ok {
				continue
			}
			results[bundleURL] = result
			if result.err != nil {
				if ed.verbose {
					log.Printf("[EndpointDiscovery] Failed to inspect %s: %v", bundleURL, result.err)
				}
				continue
			}
			for _, referenced := range result.imports {
				if len(urls) >= maxDiscoveryBundles {
					break
				}
				if _, exists := seen[referenced]; exists {
					continue
				}
				seen[referenced] = struct{}{}
				urls = append(urls, referenced)
			}
		}
		start = end
	}

	urls = prioritizeBundles(urls)
	endpoints, opFeatures := mergeBundleResults(urls, results, ed.verbose)
	return urls, endpoints, opFeatures
}

func (ed *EndpointDiscovery) processBundleWave(ctx context.Context, bundleURLs []string) map[string]bundleDiscoveryResult {
	results := make(map[string]bundleDiscoveryResult, len(bundleURLs))
	if len(bundleURLs) == 0 {
		return results
	}
	workerCount := min(bundleWorkerCount, len(bundleURLs))
	type result struct {
		url  string
		data bundleDiscoveryResult
	}
	jobs := make(chan string)
	completed := make(chan result, workerCount)
	var workers sync.WaitGroup
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case bundleURL, ok := <-jobs:
					if !ok {
						return
					}
					js, err := ed.fetchBundle(ctx, bundleURL)
					bundleResult := bundleDiscoveryResult{err: err}
					if err == nil {
						bundleResult.operations, bundleResult.features = extractOperationsFromJS(js)
						bundleResult.imports = extractReferencedBundleURLs(js, bundleURL)
					}
					select {
					case completed <- result{url: bundleURL, data: bundleResult}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, bundleURL := range bundleURLs {
			select {
			case jobs <- bundleURL:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		workers.Wait()
		close(completed)
	}()
	for item := range completed {
		results[item.url] = item.data
	}
	return results
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
	if ctx == nil {
		ctx = context.Background()
	}
	results := ed.processBundleWave(ctx, bundleURLs)
	return mergeBundleResults(bundleURLs, results, ed.verbose)
}

func mergeBundleResults(bundleURLs []string, results map[string]bundleDiscoveryResult, verbose bool) (map[string]string, map[string][]string) {
	endpoints := make(map[string]string)
	opFeatures := make(map[string][]string)
	for _, bundleURL := range bundleURLs {
		result, completed := results[bundleURL]
		if !completed {
			continue
		}
		if result.err != nil {
			if verbose {
				log.Printf("[EndpointDiscovery] Failed to extract from %s: %v", bundleURL, result.err)
			}
			continue
		}
		for opName, endpoint := range result.operations {
			if existing, ok := endpoints[opName]; ok && existing != endpoint {
				if verbose {
					log.Printf("[EndpointDiscovery] Duplicate operation %s: %s vs %s", opName, existing, endpoint)
				}
				continue
			}
			endpoints[opName] = endpoint
		}
		for opName, features := range result.features {
			if _, exists := opFeatures[opName]; !exists {
				opFeatures[opName] = append([]string(nil), features...)
			}
		}
	}
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
	if ed.client == nil {
		return "", fmt.Errorf("bundle HTTP client unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}
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

	js, err := ReadLimitedBody(resp.Body, 5*1024*1024)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

// extractOperationsFromJS parses GraphQL operations from JS content
func extractOperationsFromJS(js string) (map[string]string, map[string][]string) {
	endpoints := make(map[string]string)
	opFeatures := make(map[string][]string)

	addOperation := func(queryID, opName string) {
		decodedQueryID, err := url.PathUnescape(queryID)
		if err != nil {
			return
		}
		decodedOpName, err := url.PathUnescape(opName)
		if err != nil || decodedQueryID == "" || decodedOpName == "" {
			return
		}
		endpoints[decodedOpName] = fmt.Sprintf("%s/%s", decodedQueryID, decodedOpName)
	}

	// Some bundles expose persisted operations as request URLs instead of the
	// older queryId/operationName object shape.
	for _, match := range graphqlURLPattern.FindAllStringSubmatch(js, -1) {
		if len(match) > 2 {
			addOperation(match[1], match[2])
		}
	}

	// Vite/Rollup bundles may quote keys, reverse the record field order, or
	// construct the persisted GraphQL URL from a local query-id variable.
	for _, match := range operationRecordPattern.FindAllStringSubmatch(js, -1) {
		if len(match) > 2 {
			addOperation(match[1], match[2])
		}
	}
	for _, match := range reverseOperationRecordPattern.FindAllStringSubmatch(js, -1) {
		if len(match) > 2 {
			addOperation(match[2], match[1])
		}
	}
	queryIDs := make(map[string]string)
	for _, match := range queryIDAssignmentPattern.FindAllStringSubmatch(js, -1) {
		if len(match) > 2 {
			queryIDs[match[1]] = match[2]
		}
	}
	for _, match := range generatedGraphQLURLPattern.FindAllStringSubmatch(js, -1) {
		if len(match) > 2 {
			if queryID, ok := queryIDs[match[1]]; ok {
				addOperation(queryID, match[2])
			}
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

	// Check disk cache. Keep a recent copy if the source is temporarily blocked.
	cache, _ := ed.LoadCache()
	if cache != nil && cache.IsValid() {
		if ed.verbose {
			log.Println("[EndpointDiscovery] Using disk cache")
		}
		ed.UpdateMemoryCache(cache)
		return cache, nil
	}

	// Fetch fresh without removing the last usable cache first.
	fresh, err := ed.DiscoverEndpoints(ctx)
	if err == nil {
		return fresh, nil
	}
	if cache != nil && cache.IsUsable() {
		ed.UpdateMemoryCache(cache)
		return cache, nil
	}
	return nil, err
}

// IsUsable permits a bounded stale fallback when discovery is unavailable.
func (ec *EndpointCache) IsUsable() bool {
	return ec != nil && (len(ec.Endpoints) > 0 || len(ec.Quarantined) > 0) && !ec.Timestamp.IsZero() && time.Since(ec.Timestamp) < MaxCacheAge
}

// IsValid checks if cache is still valid (not expired)
func (ec *EndpointCache) IsValid() bool {
	if ec == nil || (len(ec.Endpoints) == 0 && len(ec.Quarantined) == 0) {
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
	if _, quarantined := ec.Quarantined[operation]; quarantined {
		return "", false
	}
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
	if err := EnsurePrivateDir(dir); err != nil {
		return err
	}

	return WriteFileAtomic(ed.cachePath, data, 0600)
}

// LoadCache loads cache from disk (public for EndpointManager)
func (ed *EndpointDiscovery) LoadCache() (*EndpointCache, error) {
	data, err := os.ReadFile(ed.cachePath)
	if err != nil {
		return nil, err
	}

	var cache EndpointCache
	if err := json.Unmarshal(data, &cache); err != nil {
		if backupErr := preserveCorruptFile(ed.cachePath); backupErr != nil {
			return nil, fmt.Errorf("invalid endpoint cache: %v; preserving corrupt cache failed: %w", err, backupErr)
		}
		return nil, fmt.Errorf("invalid endpoint cache; corrupt source preserved at %s.bak: %w", ed.cachePath, err)
	}

	return &cache, nil
}

// UpdateMemoryCache updates the global memory cache (public for EndpointManager)
func (ed *EndpointDiscovery) UpdateMemoryCache(cache *EndpointCache) {
	if cache == nil {
		return
	}
	memoryCacheMu.Lock()
	defer memoryCacheMu.Unlock()
	memoryCache = cloneEndpointCache(cache)
}

// GetMemoryCache retrieves the memory cache (public for EndpointManager)
func (ed *EndpointDiscovery) GetMemoryCache() *EndpointCache {
	cache := GetMemoryCache()
	if cache == nil || cache.Timestamp.IsZero() {
		return nil
	}
	return cache
}

// InvalidateCache clears all caches
func (ed *EndpointDiscovery) InvalidateCache() {
	memoryCacheMu.Lock()
	memoryCache = newEmptyEndpointCache()
	memoryCacheMu.Unlock()

	os.Remove(ed.cachePath)
}

// getEndpointCachePath returns the path to the endpoint cache file
func getEndpointCachePath() (string, error) {
	paths, err := GetPaths()
	if err != nil {
		return "", err
	}
	return paths.EndpointCache, nil
}

// GetDynamicGraphQLEndpoints returns endpoints from cache or fetches new ones
func GetDynamicGraphQLEndpoints() map[string]string {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return cloneStringMap(GraphQLEndpoints) // Fallback to static
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cache, err := discovery.GetCachedEndpoints(ctx)
	if err != nil {
		if Verbose {
			log.Printf("[EndpointDiscovery] Failed to get cached endpoints: %v, using static fallback", err)
		}
		return cloneStringMap(GraphQLEndpoints)
	}

	return cloneStringMap(cache.Endpoints)
}

// GetDynamicFeatures returns feature flags from cache
func GetDynamicFeatures() map[string]bool {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return cloneBoolMap(DefaultFeatures)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cache, err := discovery.GetCachedEndpoints(ctx)
	if err != nil {
		return cloneBoolMap(DefaultFeatures)
	}

	return cloneBoolMap(cache.Features)
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

	return append([]string(nil), cache.OpFeatures[operation]...)
}

func cloneStringMap(values map[string]string) map[string]string {
	copy := make(map[string]string, len(values))
	for key, value := range values {
		copy[key] = value
	}
	return copy
}

func cloneBoolMap(values map[string]bool) map[string]bool {
	copy := make(map[string]bool, len(values))
	for key, value := range values {
		copy[key] = value
	}
	return copy
}

// criticalEndpointsForCache returns health-check operations that exist in the
// current X.com bundle. X has renamed/removed the historical HomeTimeline
// operations over time, so a hard-coded check would report a false failure.
func criticalEndpointsForCache(cache *EndpointCache) []string {
	if cache == nil {
		return []string{"HomeTimeline", "Viewer", "UserByScreenName", "SearchTimeline"}
	}

	result := make([]string, 0, 5)
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

	for _, operation := range []string{"Viewer", "UserByScreenName", "SearchTimeline", "TweetDetail", "UserTweets"} {
		if _, ok := cache.Endpoints[operation]; ok {
			result = append(result, operation)
		}
	}
	return result
}

// RefreshEndpoints forces a refresh of all endpoints
func RefreshEndpoints() error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	return GetEndpointManager().RefreshEndpoints(ctx)
}

// CheckEndpointHealth checks if the current endpoints are still valid
func CheckEndpointHealth(ctx context.Context, client *XClient) (bool, []string) {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return false, []string{fmt.Sprintf("Failed to create discovery: %v", err)}
	}

	cache, err := discovery.LoadCache()

	var issues []string

	// Check if cache is getting stale
	if cache != nil && cache.IsStale() {
		issues = append(issues, "Endpoint cache is getting stale, consider refreshing")
	}
	if cache != nil {
		quarantined := make([]string, 0, len(cache.Quarantined))
		for op := range cache.Quarantined {
			quarantined = append(quarantined, op)
		}
		sort.Strings(quarantined)
		for _, op := range quarantined {
			issues = append(issues, fmt.Sprintf("%s: quarantined after a confirmed obsolete response", op))
		}
	}
	if err != nil {
		issues = append(issues, fmt.Sprintf("Dynamic cache unavailable: %v; checking static fallbacks", err))
	}
	if cache != nil && !cache.IsUsable() {
		issues = append(issues, "Dynamic cache expired; checking static fallbacks")
		cache = nil
	}

	if client == nil {
		creds := getDiscoveryCredentials()
		if creds == nil {
			issues = append(issues, "No stored authentication available for live endpoint checks")
			return false, issues
		}
		client, err = NewXClientWithRequestConfig(creds, creds.AccountName, "", RequestConfig{Timeout: 10, MaxRetries: 1, MaxResponseBytes: defaultMaxResponseBytes})
		if err != nil {
			return false, append(issues, fmt.Sprintf("Could not create endpoint probe client: %v", err))
		}
		defer client.Close()
	}
	checked := 0
	operations := criticalEndpointsForCache(cache)
	if cache != nil && len(operations) == 0 {
		operations = criticalEndpointsForCache(nil)
	}
	for _, op := range []string{"Viewer", "UserByScreenName", "SearchTimeline"} {
		found := false
		for _, candidate := range operations {
			if candidate == op {
				found = true
				break
			}
		}
		if !found {
			operations = append(operations, op)
		}
	}
	// Probe only operations with known safe inputs.
	for _, op := range operations {
		if _, _, supported := endpointProbeRequest(op); !supported {
			continue
		}
		endpoint, ok := "", false
		if cache != nil {
			if _, quarantined := cache.Quarantined[op]; quarantined {
				continue
			}
			endpoint, ok = cache.GetEndpoint(op)
		}
		if !ok {
			endpoint, ok = GraphQLEndpoints[op]
		}
		if !ok {
			issues = append(issues, fmt.Sprintf("%s: no endpoint available", op))
			continue
		}
		checked++
		probe := ProbeGraphQLEndpoint(ctx, client, op, endpoint, operationFeatures(cache, op))
		if probe.State != "healthy" {
			issues = append(issues, fmt.Sprintf("%s: %s (%s)", op, probe.Message, probe.State))
		}
	}
	if checked == 0 {
		issues = append(issues, "No operation has a safe live probe")
	}

	return len(issues) == 0, issues
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
