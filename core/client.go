package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/benoitpetit/xsh/utils"
)

// Verbose enables detailed logging of HTTP requests and responses
var Verbose = false

// logVerbose prints a message if Verbose mode is enabled
func logVerbose(format string, args ...interface{}) {
	if Verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

// APIError is raised when an API call fails
type APIError struct {
	Message      string
	StatusCode   int
	ResponseData string
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error: %s", e.Message)
}

// RateLimitError is raised when rate limited
type RateLimitError struct {
	APIError
}

// XClient is the HTTP client for Twitter/X GraphQL API
type XClient struct {
	credentials *AuthCredentials
	account     string
	proxy       string
	client      *http.Client
}

// NewXClient creates a new XClient with randomized Chrome fingerprint
func NewXClient(credentials *AuthCredentials, account, proxy string) (*XClient, error) {
	// Select a random Chrome version for this session
	chromeVersion := selectRandomChromeVersion()
	SyncChromeVersion(chromeVersion)
	
	if Verbose {
		logVerbose("Using Chrome version for TLS: %s", chromeVersion)
	}
	
	return &XClient{
		credentials: credentials,
		account:     account,
		proxy:       proxy,
	}, nil
}

// getCredentials returns cached or loaded credentials
func (c *XClient) getCredentials() (*AuthCredentials, error) {
	if c.credentials == nil {
		creds, err := GetCredentials(c.account)
		if err != nil {
			return nil, err
		}
		c.credentials = creds
	}
	return c.credentials, nil
}

// getHTTPClient returns or creates the HTTP client with uTLS fingerprinting
func (c *XClient) getHTTPClient() (*http.Client, error) {
	if c.client == nil {
		// Use uTLS for advanced TLS fingerprinting
		proxy := c.proxy
		if proxy == "" {
			proxy = os.Getenv("X_PROXY")
		}
		if proxy == "" {
			proxy = os.Getenv("TWITTER_PROXY")
		}

		client, err := newUTLSHTTPClient(proxy)
		if err != nil {
			return nil, err
		}
		c.client = client
	}
	return c.client, nil
}

// getHeaders builds request headers with Chrome impersonation
// Uses dynamic headers like Python's implementation for maximum stealth
func (c *XClient) getHeaders() (map[string]string, error) {
	return c.getHeadersWithReferer(BaseURL + "/home")
}

// getHeadersForOperation builds headers for a specific GraphQL operation
// Some operations need x-client-transaction-id (search), others must not have it (write ops)
func (c *XClient) getHeadersForOperation(operation string, referer string) (map[string]string, error) {
	// Use provided referer or default to home
	if referer == "" {
		referer = BaseURL + "/home"
	}
	
	headers, err := c.getHeadersWithReferer(referer)
	if err != nil {
		return nil, err
	}
	
	// Write operations: remove x-client-transaction-id (invalid value causes 404)
	if isWriteOperation(operation) {
		delete(headers, "x-client-transaction-id")
	} else if needsTransactionID(operation) {
		// Search operations: add x-client-transaction-id (required)
		headers["x-client-transaction-id"] = generateTransactionID()
	}
	
	return headers, nil
}

// needsTransactionID returns true for operations that require x-client-transaction-id
func needsTransactionID(operation string) bool {
	ops := map[string]bool{
		"SearchTimeline":          true,
		"HomeTimeline":            true,
		"HomeLatestTimeline":      true,
		"UserTweets":              true,
		"UserTweetsAndReplies":    true,
		"TweetDetail":             true,
		"UserByScreenName":        true,
		"Followers":               true,
		"Following":               true,
		"BookmarkSearchTimeline":  true,
	}
	return ops[operation]
}

// getHeadersWithReferer builds request headers with a custom referer
func (c *XClient) getHeadersWithReferer(referer string) (map[string]string, error) {
	creds, err := c.getCredentials()
	if err != nil {
		return nil, err
	}

	// Build dynamic headers matching the current Chrome target
	headers := map[string]string{
		"accept":                    "*/*",
		"accept-language":           GetAcceptLanguage(),
		"authorization":             "Bearer " + BearerToken,
		"content-type":              "application/json",
		"origin":                    BaseURL,
		"referer":                   referer,
		"sec-ch-ua":                 GetSecChUa(),
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        getSecChUaPlatform(),
		"sec-ch-ua-arch":            getSecChUaArch(),
		"sec-ch-ua-bitness":         "\"64\"",
		"sec-ch-ua-full-version-list": GetSecChUaFullVersionList(),
		"sec-ch-ua-model":           "\"\"",
		"sec-ch-ua-platform-version": getSecChUaPlatformVersion(),
		"sec-fetch-dest":            "empty",
		"sec-fetch-mode":            "cors",
		"sec-fetch-site":            "same-origin",
		"user-agent":                GetUserAgent(),
		"x-csrf-token":              creds.Ct0,
		"x-twitter-active-user":     "yes",
		"x-twitter-auth-type":       "OAuth2Session",
		"x-twitter-client-language": "en",
		// x-client-transaction-id may be required for search endpoints
		"x-client-transaction-id":   generateTransactionID(),
	}
	
	return headers, nil
}

// generateTransactionID generates a dummy transaction ID for x-client-transaction-id header
// This is a placeholder - the real implementation would need to reverse-engineer X.com's algorithm
func generateTransactionID() string {
	// Return a random string that looks like a transaction ID
	// Format observed: base64-like string with 90+ characters
	return "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
}

// isWriteOperation returns true for operations that don't need features in the request body
// These are typically POST operations like CreateBookmark, FavoriteTweet, etc.
func isWriteOperation(operation string) bool {
	writeOps := map[string]bool{
		"CreateBookmark":    true,
		"DeleteBookmark":    true,
		"FavoriteTweet":     true,
		"UnfavoriteTweet":   true,
		"CreateRetweet":     true,
		"DeleteRetweet":     true,
		"CreateTweet":       true,
		"CreateNoteTweet":   true,
		"DeleteTweet":       true,
		"FollowUser":        true,
		"UnfollowUser":      true,
		"CreateList":        true,
		"UpdateList":        true,
		"DeleteList":        true,
		"ListAddMember":     true,
		"ListRemoveMember":  true,
	}
	return writeOps[operation]
}

// getSecChUaPlatform returns the platform header value
func getSecChUaPlatform() string {
	return fmt.Sprintf("\"%s\"", GetPlatform())
}

// getSecChUaArch returns the architecture header value
func getSecChUaArch() string {
	return fmt.Sprintf("\"%s\"", GetArchitecture())
}

// getSecChUaPlatformVersion returns the platform version header value
func getSecChUaPlatformVersion() string {
	return fmt.Sprintf("\"%s\"", GetPlatformVersion())
}

// getCookies builds cookie map with sanitized values
func (c *XClient) getCookies() (map[string]string, error) {
	creds, err := c.getCredentials()
	if err != nil {
		return nil, err
	}

	// Use sanitized cookies to ensure valid cookie values per RFC 6265
	// This strips invalid characters like double quotes that Go's net/http rejects
	return creds.GetSanitizedCookies(), nil
}

// requestWithOperation makes an authenticated request with operation-specific headers
func (c *XClient) requestWithOperation(method, urlStr string, params, jsonData map[string]interface{}, maxRetries int, referer, operation string) (map[string]interface{}, error) {
	var headers map[string]string
	var err error
	
	if operation != "" {
		// Use operation-specific headers (removes x-client-transaction-id for write ops)
		headers, err = c.getHeadersForOperation(operation, referer)
	} else if referer != "" {
		headers, err = c.getHeadersWithReferer(referer)
	} else {
		headers, err = c.getHeaders()
	}
	if err != nil {
		return nil, err
	}

	cookies, err := c.getCookies()
	if err != nil {
		return nil, err
	}

	client, err := c.getHTTPClient()
	if err != nil {
		return nil, err
	}

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			logVerbose("Retry attempt %d/%d", attempt+1, maxRetries)
			utils.BackoffDelay(attempt, 0, 0)
		}

		var body io.Reader
		var fullURL string

		if method == "GET" && params != nil {
			// Build query string
			u, err := url.Parse(urlStr)
			if err != nil {
				return nil, err
			}
			q := u.Query()
			for k, v := range params {
				if s, ok := v.(string); ok {
					q.Set(k, s)
				} else {
					data, _ := json.Marshal(v)
					q.Set(k, string(data))
				}
			}
			u.RawQuery = q.Encode()
			fullURL = u.String()
		} else if jsonData != nil {
			data, err := json.Marshal(jsonData)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(data)
			fullURL = urlStr
			logVerbose("Request body: %s", string(data))
		} else {
			fullURL = urlStr
		}

		logVerbose("Request: %s %s", method, fullURL)

		req, err := http.NewRequest(method, fullURL, body)
		if err != nil {
			lastErr = err
			continue
		}

		// Set headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// Set cookies via Cookie header (avoids issues with Jar validation)
		var cookieParts []string
		for k, v := range cookies {
			cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", k, v))
		}
		if len(cookieParts) > 0 {
			req.Header.Set("Cookie", strings.Join(cookieParts, "; "))
		}

		if Verbose {
			logVerbose("Headers:")
			for k, v := range headers {
				logVerbose("  %s: %s", k, v)
			}
			logVerbose("Cookies (%d total):", len(cookies))
			for k, v := range cookies {
				if k == "auth_token" || k == "ct0" {
					logVerbose("  %s: %s...", k, v[:min(8, len(v))])
				} else {
					logVerbose("  %s: %s...", k, v[:min(20, len(v))])
				}
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			logVerbose("Request error: %v (type: %T)", err, err)
			lastErr = err
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		logVerbose("Response status: %d", resp.StatusCode)
		if Verbose && len(respBody) > 0 {
			// Truncate if too long
			preview := string(respBody)
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			logVerbose("Response body: %s", preview)
		}

		switch resp.StatusCode {
		case 200:
			var result map[string]interface{}
			if err := json.Unmarshal(respBody, &result); err != nil {
				logVerbose("JSON parse error: %v", err)
				logVerbose("Raw response: %s", string(respBody))
				return nil, err
			}
			return result, nil
		case 401:
			return nil, &AuthError{Message: "Authentication failed. Cookies may be expired."}
		case 429:
			if attempt < maxRetries-1 {
				continue
			}
			return nil, &RateLimitError{APIError: APIError{Message: "Rate limited by Twitter/X", StatusCode: 429}}
		case 403:
			return nil, &APIError{Message: "Forbidden — account may be suspended or action not allowed", StatusCode: 403}
		default:
			msg := string(respBody)
			if len(msg) > 500 {
				msg = msg[:500]
			}
			return nil, &APIError{Message: fmt.Sprintf("HTTP %d", resp.StatusCode), StatusCode: resp.StatusCode, ResponseData: msg}
		}
	}

	return nil, &APIError{Message: fmt.Sprintf("Request failed after %d retries: %v", maxRetries, lastErr)}
}

// GraphQLGet makes a GraphQL GET request with auto-retry on stale endpoint IDs.
// Matches Python's _graphql_request behavior exactly.
func (c *XClient) GraphQLGet(operation string, variables map[string]interface{}) (map[string]interface{}, error) {
	return c.graphqlRequest("GET", operation, variables, nil, "")
}

// GraphQLGetWithReferer makes a GraphQL GET request with a custom referer header.
func (c *XClient) GraphQLGetWithReferer(operation string, variables map[string]interface{}, referer string) (map[string]interface{}, error) {
	return c.graphqlRequest("GET", operation, variables, nil, referer)
}

// GraphQLPost makes a GraphQL POST request (for write operations) with auto-retry on stale endpoint IDs.
// Matches Python's behavior exactly: sends variables, features (from op_features), and queryId
func (c *XClient) GraphQLPost(operation string, variables map[string]interface{}) (map[string]interface{}, error) {
	return c.graphqlRequest("POST", operation, variables, nil, "")
}

// GraphQLPostWithReferer makes a GraphQL POST request with a custom referer header.
func (c *XClient) GraphQLPostWithReferer(operation string, variables map[string]interface{}, referer string) (map[string]interface{}, error) {
	return c.graphqlRequest("POST", operation, variables, nil, referer)
}

// graphqlRequest makes a GraphQL request with auto-retry on stale endpoint IDs.
// This is the internal implementation that both GraphQLGet and GraphQLPost use.
// Matches Python's _graphql_request behavior exactly.
func (c *XClient) graphqlRequest(
	method string,
	operation string,
	variables map[string]interface{},
	features map[string]bool,
	referer string,
) (map[string]interface{}, error) {
	resolvedFeatures := features
	if resolvedFeatures == nil {
		resolvedFeatures = c.getOpFeatures(operation)
	}

	// Try up to 2 times (like Python: for attempt in range(2))
	for attempt := 0; attempt < 2; attempt++ {
		// Get fresh endpoints on each attempt (after cache invalidation)
		endpoints := GetGraphQLEndpoints()
		endpoint, ok := endpoints[operation]
		if !ok {
			// Fallback to endpoint manager
			manager := GetEndpointManager()
			endpoint = manager.GetEndpoint(operation)
		}

		urlStr := GraphQLBase + "/" + endpoint
		
		if Verbose {
			logVerbose("GraphQL %s %s (attempt %d)", method, urlStr, attempt+1)
		}

		var result map[string]interface{}
		var err error

		if method == "GET" {
			params := map[string]interface{}{
				"variables":    variables,
				"features":     resolvedFeatures,
				"fieldToggles": DefaultFieldToggles,
			}
			result, err = c.requestWithOperation("GET", urlStr, params, nil, 3, referer, operation)
		} else {
			// Extract query ID from endpoint (part before /)
			queryID := endpoint
			if idx := strings.Index(endpoint, "/"); idx != -1 {
				queryID = endpoint[:idx]
			}
			
			// For write operations, don't send features (like the working curl command)
			// For read operations, send features
			var jsonData map[string]interface{}
			if isWriteOperation(operation) {
				jsonData = map[string]interface{}{
					"variables": variables,
					"queryId":   queryID,
				}
			} else {
				jsonData = map[string]interface{}{
					"variables": variables,
					"features":  resolvedFeatures,
					"queryId":   queryID,
				}
			}
			result, err = c.requestWithOperation("POST", urlStr, nil, jsonData, 3, referer, operation)
			
			if Verbose {
				jsonBytes, _ := json.Marshal(jsonData)
				logVerbose("POST body: %s", string(jsonBytes))
			}
		}

		if err != nil {
			// Check if this is a stale endpoint error (404)
			if IsEndpointObsolete(err) {
				if attempt == 0 {
					// First attempt: invalidate cache, refresh endpoints from X.com, and retry
					logVerbose("HTTP 404 for '%s' — operation IDs may be stale, "+
						"refreshing endpoints from X.com and retrying...", operation)
					InvalidateCache()
					
					// Trigger endpoint discovery to fetch fresh endpoints
					discovery, discErr := NewEndpointDiscovery(Verbose)
					if discErr == nil {
						ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
						_, discErr = discovery.DiscoverEndpoints(ctx)
						cancel()
						if discErr != nil {
							logVerbose("Failed to discover fresh endpoints: %v", discErr)
						} else {
							logVerbose("Successfully refreshed endpoints from X.com")
						}
					}

					// Refresh features if they came from cache
					if features == nil {
						resolvedFeatures = c.getOpFeatures(operation)
					}
					continue // Retry with fresh cache
				}

				// Second attempt failed too
				return nil, &APIError{
					Message: fmt.Sprintf(
						"GraphQL endpoint '%s' not found (HTTP 404) "+
							"even after cache refresh — X.com may have removed this operation",
						operation,
					),
					StatusCode:   404,
					ResponseData: err.Error(),
				}
			}
			return nil, err
		}

		// Success - apply delay like Python
		if method == "GET" {
			utils.Delay(0, 0)
		} else {
			utils.WriteDelay()
		}
		return result, nil
	}

	return nil, &APIError{
		Message: fmt.Sprintf("Unreachable: graphqlRequest retry loop for '%s'", operation),
	}
}

// getOpFeatures gets operation-specific features from cache
func (c *XClient) getOpFeatures(operation string) map[string]bool {
	opFeaturesList := GetDynamicOpFeatures(operation)
	features := make(map[string]bool)

	if opFeaturesList != nil && len(opFeaturesList) > 0 {
		cachedFeatures := GetDynamicFeatures()
		for _, feat := range opFeaturesList {
			if val, ok := cachedFeatures[feat]; ok {
				features[feat] = val
			} else {
				features[feat] = true
			}
		}
	} else {
		// Fallback to defaults
		for k, v := range DefaultFeatures {
			features[k] = v
		}
	}

	return features
}

// Close closes the HTTP client
func (c *XClient) Close() {
	if c.client != nil {
		c.client.CloseIdleConnections()
		c.client = nil
	}
}


