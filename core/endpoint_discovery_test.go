package core

import (
	"net/http"
	"reflect"
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
