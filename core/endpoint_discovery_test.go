package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDiscoveryUsesAuthenticatedHomePage(t *testing.T) {
	if HomepageURL != "https://x.com/home" {
		t.Fatalf("HomepageURL = %q, want authenticated home page", HomepageURL)
	}
}

func TestEndpointCachePathUsesPortablePaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XSH_CONFIG_DIR", root)

	paths, err := GetPaths()
	if err != nil {
		t.Fatal(err)
	}
	cachePath, err := getEndpointCachePath()
	if err != nil {
		t.Fatal(err)
	}
	if cachePath != paths.EndpointCache {
		t.Fatalf("getEndpointCachePath() = %q, want %q", cachePath, paths.EndpointCache)
	}
}

func TestApplyDiscoveryAuthAddsSessionHeadersWithoutLoggingValues(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, HomepageURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	applyDiscoveryAuth(req, &AuthCredentials{AuthToken: "auth-token-value", Ct0: "csrf-token-value"})

	if req.Header.Get("Authorization") != "Bearer "+BearerToken {
		t.Fatalf("missing bearer authorization header")
	}
	if req.Header.Get("x-csrf-token") != "csrf-token-value" {
		t.Fatalf("missing csrf header")
	}
	if req.Header.Get("x-twitter-auth-type") != "OAuth2Session" {
		t.Fatalf("missing auth type header")
	}
	if req.Header.Get("Cookie") != "auth_token=auth-token-value; ct0=csrf-token-value" {
		t.Fatalf("unexpected cookie header: %q", req.Header.Get("Cookie"))
	}
}

func TestHomepageRequestUsesBrowserCookiesWithoutAPIHeaders(t *testing.T) {
	ed := &EndpointDiscovery{credentials: &AuthCredentials{AuthToken: "auth-token-value", Ct0: "csrf-token-value"}}
	req, err := ed.newHomepageRequestAt(context.Background(), true, HomepageURL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(req.Header.Get("Cookie"), "auth_token=auth-token-value") {
		t.Fatal("homepage request omitted the browser session cookie")
	}
	for _, header := range []string{"Authorization", "x-csrf-token", "x-twitter-auth-type", "x-twitter-active-user"} {
		if req.Header.Get(header) != "" {
			t.Fatalf("homepage request sent API header %s", header)
		}
	}
}

func TestExtractBundleURLsSupportsCurrentXWebScripts(t *testing.T) {
	html := `<html><head>
<script src="https://abs.twimg.com/x-web/x-web/entry-client-logged-out-abc123.js"></script>
<script src='https://abs.twimg.com/x-web/x-web/assets/router-def456.js'></script>
</head></html>`

	got := (&EndpointDiscovery{}).extractBundleURLs(html)
	want := []string{
		"https://abs.twimg.com/x-web/x-web/entry-client-logged-out-abc123.js",
		"https://abs.twimg.com/x-web/x-web/assets/router-def456.js",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractBundleURLs() = %#v, want %#v", got, want)
	}
}

func TestExtractReferencedBundleURLsResolvesModernAssets(t *testing.T) {
	js := `const deps=["assets/router-def456.js","./chunks/timeline-abc.js"];`

	got := extractReferencedBundleURLs(js, "https://abs.twimg.com/x-web/x-web/entry.js")
	want := []string{
		"https://abs.twimg.com/x-web/x-web/assets/router-def456.js",
		"https://abs.twimg.com/x-web/x-web/chunks/timeline-abc.js",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractReferencedBundleURLs() = %#v, want %#v", got, want)
	}
}

func TestExtractOperationsFromJSSupportsMinifiedModernSyntax(t *testing.T) {
	js := "queryId:`abc_123-xyz`,operationName:`HomeTimeline_2`"

	got, _ := extractOperationsFromJS(js)
	if got["HomeTimeline_2"] != "abc_123-xyz/HomeTimeline_2" {
		t.Fatalf("operation extraction = %#v, want HomeTimeline_2 endpoint", got)
	}
}

func TestExtractOperationsFromJSParsesGraphQLURLs(t *testing.T) {
	js := `fetch("/i/api/graphql/abc-123/HomeTimeline", {method: "GET"})`

	got, _ := extractOperationsFromJS(js)
	if got["HomeTimeline"] != "abc-123/HomeTimeline" {
		t.Fatalf("URL operation extraction = %#v, want HomeTimeline endpoint", got)
	}
}

func TestCriticalEndpointSelectionUsesCurrentTimelineOperation(t *testing.T) {
	cache := &EndpointCache{Endpoints: map[string]string{
		"ConnectTabTimeline": "timeline/ConnectTabTimeline",
		"UserByScreenName":   "user/UserByScreenName",
		"SearchTimeline":     "search/SearchTimeline",
		"TweetDetail":        "tweet/TweetDetail",
		"UserTweets":         "tweets/UserTweets",
	}}

	got := criticalEndpointsForCache(cache)
	if len(got) == 0 || got[0] != "ConnectTabTimeline" {
		t.Fatalf("critical endpoint selection = %#v, want ConnectTabTimeline first", got)
	}
	for _, operation := range got {
		if operation == "HomeTimeline" || operation == "HomeLatestTimeline" {
			t.Fatalf("selected obsolete home operation %q", operation)
		}
	}
}

func TestCriticalEndpointSelectionIncludesViewerProfileCheck(t *testing.T) {
	got := criticalEndpointsForCache(&EndpointCache{
		Endpoints: map[string]string{
			"Viewer":           "viewer/Viewer",
			"UserByScreenName": "profile/UserByScreenName",
			"SearchTimeline":   "search/SearchTimeline",
		},
	})
	for _, want := range []string{"Viewer", "UserByScreenName", "SearchTimeline"} {
		found := false
		for _, operation := range got {
			if operation == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("critical endpoint list %v does not include %s", got, want)
		}
	}
}

func TestApplyDiscoveryAuthIncludesAllSanitizedStoredCookies(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, HomepageURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	applyDiscoveryAuth(req, &AuthCredentials{
		AuthToken: "auth-token-value",
		Ct0:       "csrf-token-value",
		Cookies: map[string]string{
			"guest_id":     "guest-value",
			"cf_clearance": "clearance;value",
		},
	})

	cookie := req.Header.Get("Cookie")
	for _, expected := range []string{
		"auth_token=auth-token-value",
		"ct0=csrf-token-value",
		"guest_id=guest-value",
		"cf_clearance=clearancevalue",
	} {
		if !strings.Contains(cookie, expected) {
			t.Fatalf("cookie header %q does not contain %q", cookie, expected)
		}
	}
}

func TestLoggedOutXWebShellIsNotAuthenticatedDiscoverySource(t *testing.T) {
	html := `<script type="module" src="https://abs.twimg.com/x-web/x-web/entry-client-logged-out-DjIH8Of-.js"></script>
<script>$_TSR={router:{matches:[]}}</script>`

	if !isLoggedOutShell(html) {
		t.Fatal("logged-out x-web shell was accepted as authenticated")
	}
}

func TestExpandBundleURLsTraversesNestedImportsAndCycles(t *testing.T) {
	bundles := map[string]string{}
	server := newEndpointTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		js, ok := bundles["http://"+r.Host+r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write([]byte(js))
	}))
	defer server.Close()

	entry := server.URL + "/entry.js"
	child := server.URL + "/child.js"
	grandchild := server.URL + "/grandchild.js"
	bundles[entry] = `import "./child.js"`
	bundles[child] = `import "./grandchild.js"`
	bundles[grandchild] = `import "./entry.js"`

	ed := &EndpointDiscovery{client: server.Client()}
	got := ed.expandBundleURLs(context.Background(), []string{entry})

	if len(got) != 3 {
		t.Fatalf("expandBundleURLs() returned %d bundles, want 3: %#v", len(got), got)
	}
	for _, want := range []string{entry, child, grandchild} {
		found := false
		for _, gotURL := range got {
			if gotURL == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expandBundleURLs() missing %s: %#v", want, got)
		}
	}
}

func TestDiscoveryReusesFetchedBundleBodiesForExtraction(t *testing.T) {
	const entry = "/entry.js"
	const child = "/child.js"
	var mu sync.Mutex
	requests := make(map[string]int)
	server := newEndpointTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests[r.URL.Path]++
		mu.Unlock()
		body := map[string]string{
			entry: `import "./child.js"; queryId:"entry-id",operationName:"EntryOperation"`,
			child: `queryId:"child-id",operationName:"ChildOperation"`,
		}[r.URL.Path]
		if body == "" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = io.WriteString(w, body)
	}))
	defer server.Close()

	ed := &EndpointDiscovery{client: server.Client()}
	_, endpoints, _ := ed.discoverBundleOperations(context.Background(), []string{server.URL + entry})

	if endpoints["EntryOperation"] != "entry-id/EntryOperation" || endpoints["ChildOperation"] != "child-id/ChildOperation" {
		t.Fatalf("extracted endpoints = %#v, want entry and child operations", endpoints)
	}
	mu.Lock()
	defer mu.Unlock()
	if requests[entry] != 1 || requests[child] != 1 {
		t.Fatalf("bundle requests = %#v, want one fetch per bundle", requests)
	}
}

func TestDiscoverEndpointsDoesNotPublishPartialCacheAfterCancellation(t *testing.T) {
	t.Setenv("X_AUTH_TOKEN", "test-auth-token")
	t.Setenv("X_CT0", "test-csrf-token")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cachePath := filepath.Join(t.TempDir(), "graphql_ops.json")

	memoryCacheMu.Lock()
	previousMemoryCache := cloneEndpointCache(memoryCache)
	memoryCache = newEmptyEndpointCache()
	memoryCacheMu.Unlock()
	defer func() {
		memoryCacheMu.Lock()
		memoryCache = previousMemoryCache
		memoryCacheMu.Unlock()
	}()

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var body string
		switch req.URL.Path {
		case "/home":
			body = `<script src="https://abs.twimg.com/entry.js"></script>`
		case "/entry.js":
			body = `import "./child.js"; queryId:"entry-id",operationName:"EntryOperation"`
		case "/child.js":
			cancel()
			return nil, req.Context().Err()
		default:
			return nil, fmt.Errorf("unexpected request path %q", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}
	ed := &EndpointDiscovery{client: client, cachePath: cachePath}

	_, err := ed.DiscoverEndpoints(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DiscoverEndpoints() error = %v, want context cancellation", err)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("cache file stat error = %v, want no partial cache file", err)
	}
	if cache := ed.GetMemoryCache(); cache != nil {
		t.Fatalf("partial memory cache was published: %#v", cache)
	}
}

func TestPreviewEndpointsDoesNotWriteCache(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "graphql_ops.json")
	before := GetMemoryCache()
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `queryId:"query-id",operationName:"UserByScreenName"`
		if req.URL.Path == "/home" {
			body = `<script src="https://abs.twimg.com/entry.js"></script>`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: req}, nil
	})}
	ed := &EndpointDiscovery{client: client, cachePath: cachePath, credentials: &AuthCredentials{AuthToken: "token", Ct0: "csrf"}}
	cache, err := ed.PreviewEndpoints(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cache.Endpoints["UserByScreenName"] != "query-id/UserByScreenName" {
		t.Fatalf("preview endpoints = %#v", cache.Endpoints)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("preview wrote cache: %v", err)
	}
	if !reflect.DeepEqual(GetMemoryCache(), before) {
		t.Fatal("preview changed the in-memory cache")
	}
}

func TestGetCachedEndpointsKeepsUsableCacheWhenDiscoveryFails(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "graphql_ops.json")
	previous := GetMemoryCache()
	memoryCacheMu.Lock()
	memoryCache = newEmptyEndpointCache()
	memoryCacheMu.Unlock()
	defer func() {
		memoryCacheMu.Lock()
		memoryCache = previous
		memoryCacheMu.Unlock()
	}()
	ed := &EndpointDiscovery{
		cachePath:   cachePath,
		credentials: &AuthCredentials{AuthToken: "token", Ct0: "csrf"},
		client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusUnauthorized, Body: io.NopCloser(strings.NewReader("denied")), Header: make(http.Header), Request: req}, nil
		})},
	}
	stale := &EndpointCache{Endpoints: map[string]string{"SearchTimeline": "old/SearchTimeline"}, Timestamp: time.Now().Add(-25 * time.Hour)}
	if err := ed.SaveCache(stale); err != nil {
		t.Fatal(err)
	}
	got, err := ed.GetCachedEndpoints(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Endpoints["SearchTimeline"] != "old/SearchTimeline" || !got.Timestamp.Equal(stale.Timestamp) {
		t.Fatalf("stale cache was replaced after failed discovery: %#v", got)
	}
}

func TestCheckEndpointHealthReportsKnownQuarantinedOperation(t *testing.T) {
	t.Setenv("XSH_CONFIG_DIR", t.TempDir())
	ed, err := NewEndpointDiscovery(false)
	if err != nil {
		t.Fatal(err)
	}
	cache := &EndpointCache{
		Endpoints: map[string]string{
			"UserByScreenName": "user/UserByScreenName",
			"SearchTimeline":   "search/SearchTimeline",
		},
		Quarantined: map[string]string{"Followers": "HTTP 404"},
		Timestamp:   time.Now(),
	}
	if err := ed.SaveCache(cache); err != nil {
		t.Fatal(err)
	}
	client := &XClient{requestWithOperationHook: func(_, _ string, _, _ map[string]interface{}, _ int, _, _ string) (map[string]interface{}, error) {
		return map[string]interface{}{"data": map[string]interface{}{}}, nil
	}}
	healthy, issues := CheckEndpointHealth(context.Background(), client)
	if healthy || !strings.Contains(strings.Join(issues, " "), "Followers") {
		t.Fatalf("known quarantined operation omitted: healthy=%t issues=%v", healthy, issues)
	}
}

func newEndpointTestServer(t *testing.T, handler http.Handler) (server *httptest.Server) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			if strings.Contains(strings.ToLower(recoveredString(recovered)), "operation not permitted") {
				t.Skip("sandbox does not permit local TCP listeners")
			}
			panic(recovered)
		}
	}()
	return httptest.NewServer(handler)
}

func recoveredString(value interface{}) string {
	return fmt.Sprint(value)
}

func TestExtractOperationsFromJSSupportsQuotedReversedRecords(t *testing.T) {
	js := `{"operationName":"HomeTimeline","queryId":"abc-123"}`

	got, _ := extractOperationsFromJS(js)
	if got["HomeTimeline"] != "abc-123/HomeTimeline" {
		t.Fatalf("quoted operation extraction = %#v, want normalized endpoint", got)
	}
}

func TestExtractOperationsFromJSDecodesGraphQLURLSegments(t *testing.T) {
	js := `fetch("/i/api/graphql/abc-123/HomeTimeline%5F2", {method: "GET"})`

	got, _ := extractOperationsFromJS(js)
	if got["HomeTimeline_2"] != "abc-123/HomeTimeline_2" {
		t.Fatalf("encoded URL extraction = %#v, want decoded endpoint", got)
	}
}

func TestExtractOperationsFromJSResolvesGeneratedGraphQLURL(t *testing.T) {
	js := `const queryId = "abc-123"; const url = "/i/api/graphql/" + queryId + "/HomeTimeline";`

	got, _ := extractOperationsFromJS(js)
	if got["HomeTimeline"] != "abc-123/HomeTimeline" {
		t.Fatalf("generated URL extraction = %#v, want normalized endpoint", got)
	}
}

func TestExtractOperationsFromJSIgnoresRuntimeGraphQLMetadata(t *testing.T) {
	js := `function record({queryId, operationName}) { return {queryId, operationName}; }`

	got, _ := extractOperationsFromJS(js)
	if len(got) != 0 {
		t.Fatalf("runtime metadata produced false endpoints: %#v", got)
	}
}

func TestFetchHomepageRejectsTruncatedResponse(t *testing.T) {
	ed := &EndpointDiscovery{client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", 10*1024*1024+1))), Header: make(http.Header), Request: req}, nil
	})}}
	_, _, err := ed.fetchHomepageWithClient(context.Background(), ed.client)
	var limitErr *BodyLimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf("fetchHomepageWithClient() error = %v, want BodyLimitError", err)
	}
}

func TestFetchBundleRejectsTruncatedResponse(t *testing.T) {
	ed := &EndpointDiscovery{client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", 5*1024*1024+1))), Header: make(http.Header), Request: req}, nil
	})}}
	_, err := ed.fetchBundle(context.Background(), "https://example.test/bundle.js")
	var limitErr *BodyLimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf("fetchBundle() error = %v, want BodyLimitError", err)
	}
}

func TestExtractOperationsConcurrentKeepsPrioritizedBundleOrder(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "second") {
			time.Sleep(30 * time.Millisecond)
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`queryId:"second-id",operationName:"Duplicate"`)), Header: make(http.Header), Request: req}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`queryId:"first-id",operationName:"Duplicate"`)), Header: make(http.Header), Request: req}, nil
	})}
	ed := &EndpointDiscovery{client: client}
	endpoints, _ := ed.extractOperationsConcurrent(context.Background(), []string{
		"https://example.test/first.js",
		"https://example.test/second.js",
	})
	if got := endpoints["Duplicate"]; got != "first-id/Duplicate" {
		t.Fatalf("Duplicate endpoint = %q, want prioritized first bundle", got)
	}
}

func TestExtractOperationsConcurrentStopsWorkersOnCancellation(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})}
	ed := &EndpointDiscovery{client: client}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() {
		_, _ = ed.extractOperationsConcurrent(ctx, []string{
			"https://example.test/one.js", "https://example.test/two.js", "https://example.test/three.js",
			"https://example.test/four.js", "https://example.test/five.js", "https://example.test/six.js",
		})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("extractOperationsConcurrent did not stop after cancellation")
	}
}

func TestExtractOperationsFromJSMalformedFeatureSwitchDoesNotDiscardOperation(t *testing.T) {
	js := `queryId:"abc-123",operationName:"HomeTimeline",featureSwitches:["unterminated"`
	endpoints, features := extractOperationsFromJS(js)
	if endpoints["HomeTimeline"] != "abc-123/HomeTimeline" {
		t.Fatalf("malformed feature switch discarded endpoint: %#v", endpoints)
	}
	if len(features) != 0 {
		t.Fatalf("malformed feature switch produced features: %#v", features)
	}
}
