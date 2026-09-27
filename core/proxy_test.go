package core

import (
	"errors"
	"strings"
	"testing"
)

func TestResolveProxyHonorsPrecedenceAndCredentials(t *testing.T) {
	environ := func(key string) string {
		return map[string]string{
			"CLIX_PROXY":    "http://clix.example:8080",
			"X_PROXY":       "http://x.example:8080",
			"TWITTER_PROXY": "http://twitter.example:8080",
			"HTTPS_PROXY":   "http://https.example:8080",
			"HTTP_PROXY":    "http://http.example:8080",
		}[key]
	}

	proxy, err := ResolveProxy("http://user:pass@explicit.example:3128", environ)
	if err != nil {
		t.Fatal(err)
	}
	if proxy.String() != "http://user:pass@explicit.example:3128" {
		t.Fatalf("proxy = %q, explicit proxy lost", proxy)
	}
	if proxy.User.Username() != "user" {
		t.Fatalf("proxy username = %q", proxy.User.Username())
	}

	proxy, err = ResolveProxy("", environ)
	if err != nil || proxy.String() != "http://clix.example:8080" {
		t.Fatalf("environment proxy = %v, %v; want CLIX_PROXY", proxy, err)
	}
}

func TestResolveProxyRejectsMalformedAndUnsupportedSchemes(t *testing.T) {
	for _, raw := range []string{"socks5://localhost:1080", "not a url", "http://"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ResolveProxy(raw, func(string) string { return "" }); err == nil {
				t.Fatal("ResolveProxy() accepted invalid proxy")
			}
		})
	}
}

func TestReadLimitedBodyReturnsTypedTruncationError(t *testing.T) {
	_, err := ReadLimitedBody(strings.NewReader("abcdef"), 5)
	var limitErr *BodyLimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf("ReadLimitedBody() error = %v, want BodyLimitError", err)
	}
	if limitErr.Limit != 5 {
		t.Fatalf("limit error = %#v", limitErr)
	}
}

func TestReadLimitedBodyAcceptsExactLimit(t *testing.T) {
	body, err := ReadLimitedBody(strings.NewReader("abcde"), 5)
	if err != nil || string(body) != "abcde" {
		t.Fatalf("ReadLimitedBody() = %q, %v", body, err)
	}
}
