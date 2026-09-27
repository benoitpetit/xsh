package core

import (
	"errors"
	"testing"
)

func TestRequestPolicyRetriesReadsButNotMutations(t *testing.T) {
	policy := defaultRequestPolicy()
	if !policy.CanRetry("GET", "SearchTimeline", httpStatus429, errors.New("rate limited")) {
		t.Fatal("GET rate-limit retry was rejected")
	}
	for _, tc := range []struct {
		name, method, operation string
		status                  int
		err                     error
	}{
		{name: "transport ambiguity", method: "POST", operation: "CreateTweet", err: errors.New("connection reset")},
		{name: "unprocessable mutation", method: "POST", operation: "CreateTweet", status: 422},
		{name: "server mutation", method: "POST", operation: "CreateTweet", status: 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if policy.CanRetry(tc.method, tc.operation, tc.status, tc.err) {
				t.Fatal("mutation retry was accepted")
			}
		})
	}
}

func TestGraphQLMutation422IsNotReplayed(t *testing.T) {
	client := &XClient{}
	calls := 0
	client.requestWithOperationHook = func(method, urlStr string, params, jsonData map[string]interface{}, maxRetries int, referer, operation string) (map[string]interface{}, error) {
		calls++
		return nil, &APIError{Message: "HTTP 422", StatusCode: 422}
	}

	_, err := client.graphqlRequest("POST", "CreateTweet", map[string]interface{}{"tweet_text": "hello"}, nil, "")
	if err == nil || calls != 1 {
		t.Fatalf("graphqlRequest() error=%v calls=%d, want one mutation attempt", err, calls)
	}
}

func TestIdempotentOperationAllowlistRejectsUnknownWrites(t *testing.T) {
	for _, operation := range []string{"HomeTimeline", "SearchTimeline", "TweetDetail", "UserByScreenName"} {
		if !IsIdempotentOperation(operation) {
			t.Fatalf("read operation %q is not allowlisted", operation)
		}
	}
	if IsIdempotentOperation("CreateTweet") || IsIdempotentOperation("UnknownOperation") {
		t.Fatal("write/unknown operation was allowlisted")
	}
}

const httpStatus429 = 429
