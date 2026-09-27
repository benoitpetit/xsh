package core

import (
	"fmt"
	"math"
	"net/http"
	"time"
)

const defaultMaxResponseBytes int64 = 10 * 1024 * 1024

// RequestPolicy contains validated limits for one client's request lifecycle.
type RequestPolicy struct {
	Timeout          time.Duration
	ReadAttempts     int
	MutationAttempts int
	BackoffMin       time.Duration
	BackoffMax       time.Duration
	MaxResponseBytes int64
}

var idempotentOperations = map[string]struct{}{
	"Bookmarks":             {},
	"BookmarkFoldersSlice":  {},
	"CommunityTweets":       {},
	"DMConversation":        {},
	"DMInbox":               {},
	"ExploreSidebar":        {},
	"HomeTimeline":          {},
	"JobsSearch":            {},
	"List":                  {},
	"ListMembers":           {},
	"ListTweets":            {},
	"NotificationsTimeline": {},
	"SearchTimeline":        {},
	"Trends":                {},
	"TweetDetail":           {},
	"UserByScreenName":      {},
	"UserLikes":             {},
	"UserMedia":             {},
	"UserTweets":            {},
}

// IsIdempotentOperation returns true only for the explicit read-safe set.
func IsIdempotentOperation(operation string) bool {
	_, ok := idempotentOperations[operation]
	return ok
}

// CanRetry classifies a request without replaying unknown mutations.
func (p RequestPolicy) CanRetry(method, operation string, status int, err error) bool {
	safe := method == http.MethodGet || IsIdempotentOperation(operation)
	if !safe {
		return false
	}
	if status == 0 && err != nil {
		return true
	}
	return status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= 500
}

func defaultRequestPolicy() RequestPolicy {
	return RequestPolicy{
		Timeout:          30 * time.Second,
		ReadAttempts:     3,
		MutationAttempts: 1,
		BackoffMin:       time.Second,
		BackoffMax:       30 * time.Second,
		MaxResponseBytes: defaultMaxResponseBytes,
	}
}

// RequestPolicyFromConfig validates request settings before they reach the
// transport. A zero value is intentionally invalid here; constructors apply
// compatibility defaults before calling this function.
func RequestPolicyFromConfig(config RequestConfig) (RequestPolicy, error) {
	if config.Timeout <= 0 {
		return RequestPolicy{}, fmt.Errorf("RequestConfig.Timeout must be greater than zero")
	}
	if config.MaxRetries < 1 || config.MaxRetries > 10 {
		return RequestPolicy{}, fmt.Errorf("RequestConfig.MaxRetries must be between 1 and 10")
	}
	if math.IsNaN(config.Delay) || math.IsInf(config.Delay, 0) || config.Delay < 0 {
		return RequestPolicy{}, fmt.Errorf("RequestConfig.Delay must be non-negative and finite")
	}
	if config.MaxResponseBytes <= 0 {
		return RequestPolicy{}, fmt.Errorf("RequestConfig.MaxResponseBytes must be greater than zero")
	}

	backoffMin := time.Duration(config.Delay * float64(time.Second))
	backoffMax := 30 * time.Second
	if backoffMin > backoffMax {
		backoffMax = backoffMin
	}
	return RequestPolicy{
		Timeout:          time.Duration(config.Timeout) * time.Second,
		ReadAttempts:     config.MaxRetries,
		MutationAttempts: 1,
		BackoffMin:       backoffMin,
		BackoffMax:       backoffMax,
		MaxResponseBytes: config.MaxResponseBytes,
	}, nil
}
