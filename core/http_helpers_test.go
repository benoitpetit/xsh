package core

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestAddQueryParamsEscapesAndPreservesExistingQuery(t *testing.T) {
	params := url.Values{}
	params.Set("message", "hello world")
	params.Set("tag", "a&b")
	got, err := AddQueryParams("https://example.test/path?existing=one", params)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "existing=one") || !strings.Contains(got, "message=hello+world") || !strings.Contains(got, "tag=a%26b") {
		t.Fatalf("AddQueryParams() = %q", got)
	}
}

func TestRestPostContextHonorsCancellationAnd204(t *testing.T) {
	client := &XClient{
		credentials: &AuthCredentials{AuthToken: "token", Ct0: "ct0"},
		client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			<-req.Context().Done()
			return nil, req.Context().Err()
		})},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.RestPostWithOptionsContext(ctx, "https://example.test", nil, map[string]interface{}{"ok": true}, 30)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RestPostWithOptionsContext() error = %v, want cancellation", err)
	}

	client.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header), Request: req}, nil
	})
	result, err := client.RestPostWithOptionsContext(context.Background(), "https://example.test", nil, nil, 30)
	if err != nil || len(result) != 0 {
		t.Fatalf("204 result = %#v, error=%v", result, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
