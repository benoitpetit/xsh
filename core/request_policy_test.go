package core

import (
	"strings"
	"testing"
	"time"
)

func TestRequestPolicyFromConfigValidatesBounds(t *testing.T) {
	tests := []struct {
		name string
		cfg  RequestConfig
	}{
		{name: "timeout", cfg: RequestConfig{Timeout: 0, MaxRetries: 3, MaxResponseBytes: 1}},
		{name: "attempts too low", cfg: RequestConfig{Timeout: 1, MaxRetries: 0, MaxResponseBytes: 1}},
		{name: "attempts too high", cfg: RequestConfig{Timeout: 1, MaxRetries: 11, MaxResponseBytes: 1}},
		{name: "negative delay", cfg: RequestConfig{Timeout: 1, MaxRetries: 3, Delay: -1, MaxResponseBytes: 1}},
		{name: "response limit", cfg: RequestConfig{Timeout: 1, MaxRetries: 3, MaxResponseBytes: 0}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := RequestPolicyFromConfig(tc.cfg); err == nil {
				t.Fatal("RequestPolicyFromConfig() accepted invalid config")
			}
		})
	}
}

func TestRequestPolicyFromConfigMapsConfiguredValues(t *testing.T) {
	policy, err := RequestPolicyFromConfig(RequestConfig{
		Timeout:          12,
		MaxRetries:       4,
		Delay:            0.25,
		MaxResponseBytes: 4096,
	})
	if err != nil {
		t.Fatal(err)
	}
	if policy.Timeout != 12*time.Second || policy.ReadAttempts != 4 || policy.MutationAttempts != 1 {
		t.Fatalf("unexpected policy: %#v", policy)
	}
	if policy.BackoffMin != 250*time.Millisecond || policy.MaxResponseBytes != 4096 {
		t.Fatalf("unexpected delay/limit policy: %#v", policy)
	}
}

func TestNewXClientPropagatesMaxRetries(t *testing.T) {
	client, err := NewXClientWithRequestConfig(nil, "", "", RequestConfig{
		Timeout:          12,
		MaxRetries:       7,
		MaxResponseBytes: 8192,
	})
	if err != nil {
		t.Fatal(err)
	}
	if client.policy.ReadAttempts != 7 || client.policy.MaxResponseBytes != 8192 {
		t.Fatalf("client policy = %#v, max retries/limit were not propagated", client.policy)
	}
}

func TestRequestPolicyErrorNamesField(t *testing.T) {
	_, err := RequestPolicyFromConfig(RequestConfig{Timeout: 1, MaxRetries: 11, MaxResponseBytes: 1})
	if err == nil || !strings.Contains(err.Error(), "MaxRetries") {
		t.Fatalf("error = %v, want field name", err)
	}
}
