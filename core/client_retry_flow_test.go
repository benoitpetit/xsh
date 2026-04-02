package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/benoitpetit/xsh/internal/testutil"
)

func TestGraphQLRequestRetryFlowOn404(t *testing.T) {
	client := &XClient{}
	script := testutil.NewRetryScript([]testutil.RequestStep{
		{Err: &StaleEndpointError{APIError: APIError{Message: "GraphQL endpoint not found (HTTP 404)", StatusCode: 404}}},
		{Result: map[string]interface{}{"data": map[string]interface{}{"ok": true}}},
	})
	client.requestWithOperationHook = script.RequestHook
	client.refreshEndpointsHook = script.RefreshHook
	client.invalidateCacheHook = script.InvalidateHook
	client.readDelayHook = script.ReadDelayHook

	_, err := client.graphqlRequest("GET", "SearchTimeline", map[string]interface{}{"rawQuery": "golang"}, nil, "")
	if err != nil {
		t.Fatalf("graphqlRequest returned error: %v", err)
	}

	if script.RequestCalls != 2 {
		t.Fatalf("requestCalls = %d, want 2", script.RequestCalls)
	}
	if script.RefreshCalls != 1 {
		t.Fatalf("refreshCalls = %d, want 1", script.RefreshCalls)
	}
	if script.InvalidateCalls != 1 {
		t.Fatalf("invalidateCalls = %d, want 1", script.InvalidateCalls)
	}
	if script.ReadDelayCalls != 1 {
		t.Fatalf("readDelayCalls = %d, want 1", script.ReadDelayCalls)
	}
}

func TestGraphQLRequestRetryFlowOn422(t *testing.T) {
	client := &XClient{}
	script := testutil.NewRetryScript([]testutil.RequestStep{
		{Err: &APIError{Message: "HTTP 422", StatusCode: 422}},
		{Result: map[string]interface{}{"data": map[string]interface{}{"ok": true}}},
	})
	client.requestWithOperationHook = script.RequestHook
	client.refreshEndpointsHook = script.RefreshHook
	client.invalidateCacheHook = script.InvalidateHook
	client.writeDelayHook = script.WriteDelayHook

	_, err := client.graphqlRequest("POST", "CreateTweet", map[string]interface{}{"tweet_text": "hello"}, nil, "")
	if err != nil {
		t.Fatalf("graphqlRequest returned error: %v", err)
	}

	if script.RequestCalls != 2 {
		t.Fatalf("requestCalls = %d, want 2", script.RequestCalls)
	}
	if script.RefreshCalls != 1 {
		t.Fatalf("refreshCalls = %d, want 1", script.RefreshCalls)
	}
	if script.InvalidateCalls != 1 {
		t.Fatalf("invalidateCalls = %d, want 1", script.InvalidateCalls)
	}
	if script.WriteDelayCalls != 1 {
		t.Fatalf("writeDelayCalls = %d, want 1", script.WriteDelayCalls)
	}
}

func TestGraphQLRequestRetryFlowOnBodyStaleError(t *testing.T) {
	client := &XClient{}
	script := testutil.NewRetryScript([]testutil.RequestStep{
		{
			Result: map[string]interface{}{
				"errors": []interface{}{
					map[string]interface{}{"message": "Query not found"},
				},
			},
		},
		{Result: map[string]interface{}{"data": map[string]interface{}{"ok": true}}},
	})
	client.requestWithOperationHook = script.RequestHook
	client.refreshEndpointsHook = script.RefreshHook
	client.invalidateCacheHook = script.InvalidateHook
	client.readDelayHook = script.ReadDelayHook

	_, err := client.graphqlRequest("GET", "UserTweets", map[string]interface{}{"userId": "123"}, nil, "")
	if err != nil {
		t.Fatalf("graphqlRequest returned error: %v", err)
	}

	if script.RequestCalls != 2 {
		t.Fatalf("requestCalls = %d, want 2", script.RequestCalls)
	}
	if script.RefreshCalls != 1 {
		t.Fatalf("refreshCalls = %d, want 1", script.RefreshCalls)
	}
	if script.InvalidateCalls != 1 {
		t.Fatalf("invalidateCalls = %d, want 1", script.InvalidateCalls)
	}
	if script.ReadDelayCalls != 1 {
		t.Fatalf("readDelayCalls = %d, want 1", script.ReadDelayCalls)
	}
}

func TestGraphQLRequestReturnsErrorAfterRetryExhausted(t *testing.T) {
	client := &XClient{}
	script := testutil.NewRetryScript([]testutil.RequestStep{
		{Err: &StaleEndpointError{APIError: APIError{Message: "GraphQL endpoint not found (HTTP 404)", StatusCode: 404}}},
		{Err: &StaleEndpointError{APIError: APIError{Message: "GraphQL endpoint not found (HTTP 404)", StatusCode: 404}}},
	})
	script.RefreshErr = errors.New("refresh failed")
	client.requestWithOperationHook = script.RequestHook
	client.refreshEndpointsHook = script.RefreshHook
	client.invalidateCacheHook = script.InvalidateHook

	_, err := client.graphqlRequest("GET", "TweetDetail", map[string]interface{}{"focalTweetId": "1"}, nil, "")
	if err == nil {
		t.Fatalf("expected graphqlRequest to return an error")
	}

	if script.RequestCalls != 2 {
		t.Fatalf("requestCalls = %d, want 2", script.RequestCalls)
	}
	if script.RefreshCalls != 1 {
		t.Fatalf("refreshCalls = %d, want 1", script.RefreshCalls)
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Fatalf("apiErr.StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if !strings.Contains(strings.ToLower(apiErr.Message), "not found") {
		t.Fatalf("unexpected APIError message: %q", apiErr.Message)
	}
}
