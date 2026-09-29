package core

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/benoitpetit/xsh/internal/testutil"
)

func TestGraphQLRequestRetryFlowOn404(t *testing.T) {
	resetEndpointTestState()
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

func TestGraphQLRetryPassesClientContextToEndpointRefresh(t *testing.T) {
	resetEndpointTestState()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := &XClient{}
	client.SetContext(ctx)
	var requests, refreshes int
	client.requestWithOperationHook = func(_, _ string, _, _ map[string]interface{}, _ int, _, _ string) (map[string]interface{}, error) {
		requests++
		return nil, &StaleEndpointError{APIError: APIError{Message: "stale endpoint", StatusCode: 404}}
	}
	client.refreshEndpointsContextHook = func(got context.Context, _ string) error {
		refreshes++
		if got != ctx {
			t.Fatal("refresh did not receive the command context")
		}
		cancel()
		return got.Err()
	}

	_, err := client.GraphQLGet("SearchTimeline", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GraphQL error = %v, want context cancellation", err)
	}
	if requests != 1 || refreshes != 1 {
		t.Fatalf("requests/refreshes = %d/%d, want 1/1", requests, refreshes)
	}
}

func TestRateLimitWaitHonorsClientContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &XClient{}
	client.SetContext(ctx)
	cancel()

	started := time.Now()
	err := client.waitForRetryAfter(time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("rate limit wait error = %v, want context cancellation", err)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("cancelled rate limit wait took %s, want prompt return", elapsed)
	}
}

func resetEndpointTestState() {
	memoryCacheMu.Lock()
	memoryCache = newEmptyEndpointCache()
	memoryCacheMu.Unlock()
	manager := GetEndpointManager()
	manager.repositoryMu.Lock()
	repository := NewEndpointRepository(nil)
	_ = repository.Replace(&EndpointCache{
		Endpoints: map[string]string{"TestOnly": "test/TestOnly"},
		Timestamp: time.Now(),
	})
	manager.repository = repository
	manager.repositoryMu.Unlock()
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
	if err == nil {
		t.Fatal("graphqlRequest unexpectedly replayed a mutation after HTTP 422")
	}

	if script.RequestCalls != 1 {
		t.Fatalf("requestCalls = %d, want 1", script.RequestCalls)
	}
	if script.RefreshCalls != 0 {
		t.Fatalf("refreshCalls = %d, want 0", script.RefreshCalls)
	}
	if script.InvalidateCalls != 0 {
		t.Fatalf("invalidateCalls = %d, want 0", script.InvalidateCalls)
	}
	if script.WriteDelayCalls != 0 {
		t.Fatalf("writeDelayCalls = %d, want 0", script.WriteDelayCalls)
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

func TestGraphQLRequestQuarantinesEndpointAfterRefreshReturnsSameBrokenID(t *testing.T) {
	manager := GetEndpointManager()
	manager.repositoryMu.Lock()
	previousRepository := manager.repository
	manager.repository = NewEndpointRepository(nil)
	manager.repositoryMu.Unlock()
	previousMemory := GetMemoryCache()
	memoryCacheMu.Lock()
	memoryCache = newEmptyEndpointCache()
	memoryCacheMu.Unlock()
	t.Cleanup(func() {
		manager.repositoryMu.Lock()
		manager.repository = previousRepository
		manager.repositoryMu.Unlock()
		memoryCacheMu.Lock()
		memoryCache = previousMemory
		memoryCacheMu.Unlock()
	})
	manager.UpdateEndpoint("Followers", "old/Followers")

	requests := 0
	client := &XClient{requestWithOperationHook: func(_, _ string, _, _ map[string]interface{}, _ int, _, _ string) (map[string]interface{}, error) {
		requests++
		return nil, &StaleEndpointError{APIError: APIError{StatusCode: 404}}
	}}
	client.refreshEndpointsContextHook = func(context.Context, string) error {
		manager.UpdateEndpoint("Followers", "rediscovered/Followers")
		return nil
	}
	_, err := client.GraphQLGet("Followers", map[string]interface{}{"userId": "123", "count": 1})
	if err == nil || requests != 2 {
		t.Fatalf("GraphQL request error=%v requests=%d, want two failures", err, requests)
	}
	if !manager.IsQuarantined("Followers") {
		t.Fatal("endpoint remained active after both IDs returned 404")
	}
}
