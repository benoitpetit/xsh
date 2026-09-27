package core

import (
	"testing"
	"time"
)

func TestRateLimitCopiesProtectGlobalState(t *testing.T) {
	endpoint := "copy-test"
	original := &RateLimitInfo{Endpoint: endpoint, Limit: 100, Remaining: 50, Reset: time.Now()}
	rateLimitStore.Update(endpoint, original)

	original.Remaining = 1
	got := GetRateLimit(endpoint)
	if got == nil || got.Remaining != 50 {
		t.Fatalf("stored rate limit = %#v, input mutation leaked", got)
	}
	got.Remaining = 2
	if again := GetRateLimit(endpoint); again.Remaining != 50 {
		t.Fatalf("returned rate limit mutation leaked: %#v", again)
	}
}

func TestRateLimitUsagePercentIsClamped(t *testing.T) {
	if got := (&RateLimitInfo{Limit: 10, Remaining: -5}).UsagePercent(); got != 100 {
		t.Fatalf("UsagePercent() = %v, want 100", got)
	}
	if got := (&RateLimitInfo{Limit: 10, Remaining: 20}).UsagePercent(); got != 0 {
		t.Fatalf("UsagePercent() = %v, want 0", got)
	}
}

func TestParseRetryAfterSupportsSecondsAndHTTPDate(t *testing.T) {
	if got := parseRetryAfter("7", time.Time{}); got != 7*time.Second {
		t.Fatalf("seconds Retry-After = %s, want 7s", got)
	}
	now := time.Now().UTC()
	got := parseRetryAfter(now.Add(5*time.Second).Format(httpTimeFormat), now)
	if got < 4*time.Second || got > 6*time.Second {
		t.Fatalf("date Retry-After = %s, want about 5s", got)
	}
}
