package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoveryUsesAuthenticatedHomePage(t *testing.T) {
	if HomepageURL != "https://x.com/home" {
		t.Fatalf("HomepageURL = %q, want authenticated home page", HomepageURL)
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
