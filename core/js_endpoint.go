// Package core provides endpoint discovery from X.com JavaScript bundles.
// This approach extracts GraphQL operation IDs directly from the public JS bundles,
// which doesn't require authentication.
package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	xHomepageURL      = "https://x.com/elonmusk"
	bundleCDNBase     = "https://abs.twimg.com/responsive-web/client-web"
	endpointCacheFile = "graphql_ops.json"
	cacheTTL          = 24 * time.Hour
)

// Regex patterns are defined in endpoint_discovery.go

// EndpointData holds discovered endpoints and features
type EndpointData struct {
	Endpoints   map[string]string
	Features    map[string]bool
	OpFeatures  map[string][]string
}

// JSEndpointDiscovery discovers endpoints by parsing X.com JS bundles
type JSEndpointDiscovery struct {
	cachePath string
}

// NewJSEndpointDiscovery creates a new JS-based endpoint discovery
func NewJSEndpointDiscovery() *JSEndpointDiscovery {
	cachePath := getJSCachePath()
	return &JSEndpointDiscovery{
		cachePath: cachePath,
	}
}

// DiscoverEndpoints fetches X.com JS bundles and extracts GraphQL operation IDs
func (jsd *JSEndpointDiscovery) DiscoverEndpoints() (map[string]string, error) {
	data, err := jsd.DiscoverAll()
	if err != nil {
		return nil, err
	}
	return data.Endpoints, nil
}

// DiscoverAll fetches endpoints and features from X.com
func (jsd *JSEndpointDiscovery) DiscoverAll() (*EndpointData, error) {
	fmt.Println("🔍 Discovering endpoints from X.com JavaScript bundles...")
	fmt.Println()

	if cached := jsd.LoadFullCache(); cached != nil {
		fmt.Println("✅ Using cached endpoints (less than 24h old)")
		return cached, nil
	}

	fmt.Println("📡 Fetching X.com homepage...")
	html, err := jsd.fetchHomepage()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch homepage: %w", err)
	}
	fmt.Println("✅ Homepage fetched")
	fmt.Println()

	// Extract features from HTML (window.__INITIAL_STATE__)
	fmt.Println("🔧 Extracting feature switches...")
	features := jsd.extractFeaturesFromHTML(html)
	fmt.Printf("✅ Found %d feature switches\n", len(features))
	fmt.Println()

	fmt.Println("📦 Extracting JS bundle URLs...")
	bundleURLs := jsd.extractBundleURLs(html)
	if len(bundleURLs) == 0 {
		return nil, fmt.Errorf("no JS bundle URLs found in homepage")
	}
	fmt.Printf("✅ Found %d JS bundles\n", len(bundleURLs))
	fmt.Println()

	fmt.Println("🔎 Parsing JS bundles for GraphQL operations...")
	endpoints := make(map[string]string)
	opFeatures := make(map[string][]string)

	for i, url := range bundleURLs {
		fmt.Printf("   [%d/%d] Downloading %s... ", i+1, len(bundleURLs), jsd.shortURL(url))

		jsContent, err := jsd.fetchBundle(url)
		if err != nil {
			fmt.Printf("⚠️  skipped (%v)\n", err)
			continue
		}

		ops, opFeats := jsd.extractOperations(jsContent)
		if len(ops) > 0 {
			fmt.Printf("✓ found %d operations\n", len(ops))
			for name, endpoint := range ops {
				endpoints[name] = endpoint
			}
			for name, feats := range opFeats {
				opFeatures[name] = feats
			}
		} else {
			fmt.Printf("- no operations\n")
		}
	}

	fmt.Println()

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no GraphQL operations found in any bundle")
	}

	fmt.Printf("✅ Discovered %d GraphQL endpoints\n", len(endpoints))

	data := &EndpointData{
		Endpoints:  endpoints,
		Features:   features,
		OpFeatures: opFeatures,
	}

	jsd.saveFullCache(data)

	return data, nil
}

func (jsd *JSEndpointDiscovery) fetchHomepage() (string, error) {
	// Use uTLS client for browser fingerprinting (same as endpoint_discovery.go)
	client, err := newUTLSHTTPClient("")
	if err != nil {
		// Fallback to standard client
		client = &http.Client{Timeout: 15 * time.Second}
	}

	req, err := http.NewRequest("GET", xHomepageURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("user-agent", GetUserAgent())
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("accept-language", GetAcceptLanguage())
	// Note: Don't set Accept-Encoding to avoid compression issues

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// extractFeaturesFromHTML extracts feature flags from window.__INITIAL_STATE__
func (jsd *JSEndpointDiscovery) extractFeaturesFromHTML(html string) map[string]bool {
	features := make(map[string]bool)
	
	marker := "window.__INITIAL_STATE__="
	idx := strings.Index(html, marker)
	if idx == -1 {
		return features
	}
	
	jsonStart := idx + len(marker)
	jsonStr := jsd.extractJSONObject(html, jsonStart)
	if jsonStr == "" {
		return features
	}
	
	var state map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &state); err != nil {
		return features
	}
	
	// Navigate to featureSwitch.defaultConfig
	if fs, ok := state["featureSwitch"].(map[string]interface{}); ok {
		var config map[string]interface{}
		if cfg, ok := fs["defaultConfig"].(map[string]interface{}); ok {
			config = cfg
		} else if cfg, ok := fs["features"].(map[string]interface{}); ok {
			config = cfg
		}
		
		for key, val := range config {
			if v, ok := val.(map[string]interface{}); ok {
				if val, ok := v["value"].(bool); ok {
					features[key] = val
				}
			} else if val, ok := val.(bool); ok {
				features[key] = val
			}
		}
	}
	
	return features
}

// extractJSONObject extracts a complete JSON object using brace counting
func (jsd *JSEndpointDiscovery) extractJSONObject(text string, start int) string {
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

func (jsd *JSEndpointDiscovery) extractBundleURLs(html string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, match := range bundleHrefPattern.FindAllStringSubmatch(html, -1) {
		if len(match) > 1 {
			url := match[1]
			if !seen[url] {
				seen[url] = true
				result = append(result, url)
			}
		}
	}

	for _, match := range bundleSrcPattern.FindAllStringSubmatch(html, -1) {
		if len(match) > 1 {
			url := match[1]
			if !seen[url] {
				seen[url] = true
				result = append(result, url)
			}
		}
	}

	for _, match := range chunkMapPattern.FindAllStringSubmatch(html, -1) {
		if len(match) > 1 {
			var chunkMap map[string]string
			if err := json.Unmarshal([]byte(match[1]), &chunkMap); err == nil {
				for name, hash := range chunkMap {
					url := fmt.Sprintf("%s/%s.%sa.js", bundleCDNBase, name, hash)
					if !seen[url] {
						seen[url] = true
						result = append(result, url)
					}
				}
			}
		}
	}

	var mainBundles, otherBundles []string
	for _, url := range result {
		if strings.Contains(url, "/main.") {
			mainBundles = append(mainBundles, url)
		} else {
			otherBundles = append(otherBundles, url)
		}
	}

	return append(mainBundles, otherBundles...)
}

func (jsd *JSEndpointDiscovery) fetchBundle(url string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")
	req.Header.Set("accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// extractOperations extracts GraphQL operations and their feature switches from JS
func (jsd *JSEndpointDiscovery) extractOperations(jsContent string) (map[string]string, map[string][]string) {
	ops := make(map[string]string)
	opFeatures := make(map[string][]string)

	// Find all operation patterns
	// Look for queryId:"XXX",operationName:"YYY"
	pattern := regexp.MustCompile(`queryId:\s*"([A-Za-z0-9_-]+)"[^}]*?operationName:\s*"([A-Za-z]+)"`)
	matches := pattern.FindAllStringSubmatchIndex(jsContent, -1)

	for i, match := range matches {
		if len(match) < 6 {
			continue
		}

		// Extract queryId and operationName from the match
		queryID := jsContent[match[2]:match[3]]
		opName := jsContent[match[4]:match[5]]
		endpoint := fmt.Sprintf("%s/%s", queryID, opName)

		if _, exists := ops[opName]; !exists {
			ops[opName] = endpoint

			// Extract the block for this operation (from this match to the next)
			blockEnd := len(jsContent)
			if i+1 < len(matches) {
				blockEnd = matches[i+1][0]
			}
			block := jsContent[match[0]:blockEnd]

			// Extract feature switches for this operation
			featureSwitchesPattern := regexp.MustCompile(`featureSwitches:\s*(\[[^\]]*\])`)
			if fsMatch := featureSwitchesPattern.FindStringSubmatch(block); fsMatch != nil {
				var features []string
				if err := json.Unmarshal([]byte(fsMatch[1]), &features); err == nil {
					opFeatures[opName] = features
				}
			}
		}
	}

	return ops, opFeatures
}


func (jsd *JSEndpointDiscovery) shortURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return url
}

// LoadFullCache loads the full cache from disk if valid
func (jsd *JSEndpointDiscovery) LoadFullCache() *EndpointData {
	data, err := os.ReadFile(jsd.cachePath)
	if err != nil {
		return nil
	}

	var cache struct {
		Endpoints  map[string]string   `json:"endpoints"`
		Features   map[string]bool     `json:"features"`
		OpFeatures map[string][]string `json:"op_features"`
		Timestamp  time.Time           `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}

	age := time.Since(cache.Timestamp)
	if age > cacheTTL {
		return nil
	}

	return &EndpointData{
		Endpoints:  cache.Endpoints,
		Features:   cache.Features,
		OpFeatures: cache.OpFeatures,
	}
}

func (jsd *JSEndpointDiscovery) saveFullCache(data *EndpointData) {
	cache := struct {
		Endpoints  map[string]string   `json:"endpoints"`
		Features   map[string]bool     `json:"features"`
		OpFeatures map[string][]string `json:"op_features"`
		Timestamp  time.Time           `json:"timestamp"`
	}{
		Endpoints:  data.Endpoints,
		Features:   data.Features,
		OpFeatures: data.OpFeatures,
		Timestamp:  time.Now(),
	}

	jsonData, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return
	}

	dir := filepath.Dir(jsd.cachePath)
	os.MkdirAll(dir, 0755)

	os.WriteFile(jsd.cachePath, jsonData, 0600)
}

func getJSCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", ConfigDirName, endpointCacheFile)
}

// invalidateJSCache clears the endpoint cache
// UpdateObsoleteEndpoints checks all endpoints and updates obsolete ones
func (jsd *JSEndpointDiscovery) UpdateObsoleteEndpoints() error {
	manager := GetEndpointManager()

	fmt.Println("\n🔍 Checking all endpoints for obsolescence...")

	// Quick check of critical endpoints (not all, to save time)
	endpointsToCheck := []string{
		"HomeTimeline",
		"UserByScreenName",
		"SearchTimeline",
		"TweetDetail",
		"CreateTweet",
		"UserTweets",
		"Followers",
		"Following",
		"Bookmarks",
	}

	obsoleteEndpoints := []string{}

	for _, operation := range endpointsToCheck {
		fmt.Printf("   Checking %s... ", operation)

		// Try to make a minimal request to check if endpoint works
		if err := jsd.checkEndpointWithClient(operation); err != nil {
			fmt.Printf("❌ OBSOLETE\n")
			obsoleteEndpoints = append(obsoleteEndpoints, operation)
		} else {
			fmt.Printf("✅ OK\n")
		}
	}

	if len(obsoleteEndpoints) == 0 {
		fmt.Println("\n✅ All endpoints are up to date!")
		return nil
	}

	fmt.Printf("\n⚠️  Found %d obsolete endpoints: %v\n", len(obsoleteEndpoints), obsoleteEndpoints)

	// Try JS-based discovery (new approach - doesn't require auth)
	fmt.Println("🔄 Starting auto-discovery from X.com JS bundles...")
	fmt.Println("   (This extracts endpoint IDs from public JavaScript bundles)")
	fmt.Println()

	discovered, err := jsd.DiscoverEndpoints()
	if err != nil {
		fmt.Printf("\n⚠️  Auto-discovery failed: %v\n", err)
		fmt.Println()
		return jsd.manualUpdatePrompt(obsoleteEndpoints)
	}

	// Update obsolete endpoints with discovered ones
	updated := 0
	for _, operation := range obsoleteEndpoints {
		if newEndpoint, ok := discovered[operation]; ok {
			manager.UpdateEndpoint(operation, newEndpoint)
			fmt.Printf("✅ Updated %s -> %s\n", operation, newEndpoint)
			updated++
		} else {
			fmt.Printf("⚠️  Could not discover new endpoint for %s\n", operation)
		}
	}

	fmt.Printf("\n✅ Updated %d/%d obsolete endpoints\n", updated, len(obsoleteEndpoints))

	if updated < len(obsoleteEndpoints) {
		// Some endpoints couldn't be discovered automatically
		missing := []string{}
		for _, op := range obsoleteEndpoints {
			if _, ok := discovered[op]; !ok {
				missing = append(missing, op)
			}
		}
		return jsd.manualUpdatePrompt(missing)
	}

	return nil
}

// manualUpdatePrompt shows instructions for manual endpoint update
func (jsd *JSEndpointDiscovery) manualUpdatePrompt(obsoleteEndpoints []string) error {
	fmt.Println()
	fmt.Println("┌──────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│  🔧 MANUAL UPDATE REQUIRED                                               │")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│                                                                          │")
	fmt.Println("│  Auto-discovery requires you to be logged into x.com in your browser.    │")
	fmt.Println("│                                                                          │")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│  ⚡ QUICK FIX: Find new endpoints in 30 seconds                         │")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│                                                                          │")
	fmt.Println("│  1. Open Chrome/Brave/Edge and go to https://x.com                       │")
	fmt.Println("│     (make sure you're LOGGED IN)                                         │")
	fmt.Println("│                                                                          │")
	fmt.Println("│  2. Press F12 → Click 'Network' tab → Type 'graphql' in filter           │")
	fmt.Println("│                                                                          │")
	fmt.Println("│  3. Navigate to trigger the obsolete endpoints:                          │")

	for _, op := range obsoleteEndpoints {
		desc := ""
		switch op {
		case "UserByScreenName":
			desc = " (click any profile)"
		case "SearchTimeline":
			desc = " (search something)"
		case "Followers":
			desc = " (view followers)"
		case "UserTweets":
			desc = " (view a profile)"
		case "TweetDetail":
			desc = " (click a tweet)"
		}
		fmt.Printf("│     • %s%s\n", op, desc)
	}

	fmt.Println("│                                                                          │")
	fmt.Println("│  4. Look for requests containing these names, you'll see:                │")
	fmt.Println("│     graphql/NEW_ID/OperationName                                         │")
	fmt.Println("│                                                                          │")
	fmt.Println("│  5. Copy the NEW_ID and update:                                          │")
	fmt.Println("│     xsh endpoints update UserByScreenName 'NEW_ID/UserByScreenName'     │")
	fmt.Println("│                                                                          │")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│  💡 ALTERNATIVE: Community endpoints                                     │")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│                                                                          │")
	fmt.Println("│  Check if updated endpoints are shared here:                             │")
	fmt.Println("│  https://github.com/benoitpetit/xsh/issues                              │")
	fmt.Println("│                                                                          │")
	fmt.Println("│  Or try these commonly working IDs (may change):                         │")
	fmt.Println("│  • UserByScreenName: G8VeVwn5NiyWfYJ2xNIvQQ or vNixRVrCfoEHf1C9fbHDlQ   │")
	fmt.Println("│  • SearchTimeline:   BbGLL1ZfMQ2jogw0UCiNkg or 7jzx5IlQp3eoa7t2HjJWnw   │")
	fmt.Println("│                                                                          │")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│  📋 EXAMPLE COMMANDS                                                     │")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│                                                                          │")
	for _, op := range obsoleteEndpoints {
		fmt.Printf("│  xsh endpoints update %s 'NEW_ID/%s'\n", op, op)
	}
	fmt.Println("│                                                                          │")
	fmt.Printf("│  Then verify: xsh endpoints check %s                                    │\n", obsoleteEndpoints[0])
	fmt.Println("│                                                                          │")
	fmt.Println("└──────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()

	return nil
}

// checkEndpointWithClient makes a test request to check if endpoint works
func (jsd *JSEndpointDiscovery) checkEndpointWithClient(operation string) error {
	// Create a temporary client
	client, err := NewXClient(nil, "", "")
	if err != nil {
		return err
	}
	defer client.Close()

	// Try a minimal request based on operation type
	var variables map[string]interface{}

	switch operation {
	case "UserByScreenName":
		variables = map[string]interface{}{"screen_name": "twitter"}
	case "SearchTimeline":
		variables = map[string]interface{}{
			"rawQuery":    "test",
			"count":       1,
			"querySource": "typed_query",
			"product":     "Top",
		}
	case "TweetDetail":
		variables = map[string]interface{}{"focalTweetId": "12345"}
	default:
		variables = map[string]interface{}{"count": 1}
	}

	_, err = client.GraphQLGet(operation, variables)

	if err != nil {
		// Check if it's a 404 "Query not found" error
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("endpoint obsolete")
		}
		// Other errors (401, 403, 400) mean endpoint exists
		return nil
	}

	return nil
}
