// Package core provides rate limit tracking for Twitter/X API.
package core

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimitInfo holds rate limit data from response headers
type RateLimitInfo struct {
	Endpoint  string    `json:"endpoint"`
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SecondsUntilReset returns seconds until the rate limit resets
func (r *RateLimitInfo) SecondsUntilReset() int {
	d := time.Until(r.Reset).Seconds()
	if d < 0 {
		return 0
	}
	return int(d)
}

// UsagePercent returns the percentage of rate limit used
func (r *RateLimitInfo) UsagePercent() float64 {
	if r.Limit == 0 {
		return 0
	}
	usage := float64(r.Limit-r.Remaining) / float64(r.Limit) * 100
	if usage < 0 {
		return 0
	}
	if usage > 100 {
		return 100
	}
	return usage
}

// rateLimitStore is a global in-memory store for rate limit data
var rateLimitStore = &rateLimitMap{
	data: make(map[string]*RateLimitInfo),
}

type rateLimitMap struct {
	mu   sync.RWMutex
	data map[string]*RateLimitInfo
}

// UpdateRateLimit stores rate limit info for an endpoint
func (m *rateLimitMap) Update(endpoint string, info *RateLimitInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[endpoint] = cloneRateLimitInfo(info)
}

// GetRateLimit returns rate limit info for an endpoint
func (m *rateLimitMap) Get(endpoint string) *RateLimitInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneRateLimitInfo(m.data[endpoint])
}

// GetAll returns all tracked rate limits
func (m *rateLimitMap) GetAll() []*RateLimitInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*RateLimitInfo, 0, len(m.data))
	for _, v := range m.data {
		result = append(result, cloneRateLimitInfo(v))
	}
	return result
}

// GetRateLimits returns all tracked rate limit info (exported for cmd layer)
func GetRateLimits() []*RateLimitInfo {
	return rateLimitStore.GetAll()
}

// GetRateLimit returns a defensive copy of one endpoint's state.
func GetRateLimit(endpoint string) *RateLimitInfo {
	return rateLimitStore.Get(endpoint)
}

func cloneRateLimitInfo(info *RateLimitInfo) *RateLimitInfo {
	if info == nil {
		return nil
	}
	copy := *info
	return &copy
}

const httpTimeFormat = http.TimeFormat

func parseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0
	}
	if now.IsZero() {
		now = time.Now()
	}
	if delay := time.Until(when); !now.IsZero() {
		delay = when.Sub(now)
		if delay < 0 {
			return 0
		}
		return delay
	}
	return 0
}

func parseRateLimitReset(value string) time.Time {
	seconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || seconds <= 0 {
		return time.Time{}
	}
	return time.Unix(seconds, 0)
}
