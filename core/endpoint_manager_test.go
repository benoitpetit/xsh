package core

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestEndpointStatsDistinguishesStaticAndDynamicEndpoints(t *testing.T) {
	manager := &EndpointManager{}

	stats := manager.GetStats()

	if stats.StaticCount != len(GraphQLEndpoints) {
		t.Fatalf("StaticCount = %d, want %d", stats.StaticCount, len(GraphQLEndpoints))
	}
	if stats.DynamicCount != 0 {
		t.Fatalf("DynamicCount = %d, want 0", stats.DynamicCount)
	}
	if stats.TotalCount != len(GraphQLEndpoints) {
		t.Fatalf("TotalCount = %d, want %d", stats.TotalCount, len(GraphQLEndpoints))
	}
	if stats.CacheAge != 0 {
		t.Fatalf("CacheAge = %s, want 0 for an empty cache", stats.CacheAge)
	}
}

func TestRefreshForOperationUsesCooldownAfterFailure(t *testing.T) {
	var refreshes int
	manager := &EndpointManager{
		refreshCooldown: time.Minute,
		refreshFunc: func(context.Context) error {
			refreshes++
			return errors.New("discovery unavailable")
		},
	}

	firstErr := manager.RefreshForOperation(context.Background(), "MissingOperation")
	secondErr := manager.RefreshForOperation(context.Background(), "MissingOperation")
	if firstErr == nil || secondErr == nil {
		t.Fatal("expected both refresh attempts to return an error")
	}
	if refreshes != 1 {
		t.Fatalf("refreshes = %d, want 1 during cooldown", refreshes)
	}
}

func TestRefreshForOperationPreservesExistingEndpointOnFailure(t *testing.T) {
	operation := "PreservedOperation"
	discovery := &EndpointDiscovery{}
	discovery.UpdateMemoryCache(&EndpointCache{
		Endpoints: map[string]string{operation: "old-query-id/old-operation"},
		Features:  map[string]bool{}, OpFeatures: map[string][]string{}, Timestamp: time.Now(),
	})
	manager := &EndpointManager{
		discovery: discovery,
		refreshFunc: func(context.Context) error {
			return errors.New("discovery unavailable")
		},
	}

	if err := manager.RefreshForOperation(context.Background(), operation); err == nil {
		t.Fatal("expected refresh to fail")
	}
	if got := manager.GetEndpoint(operation); got != "old-query-id/old-operation" {
		t.Fatalf("endpoint after failed refresh = %q, want previous endpoint", got)
	}
}

func TestQuarantineEndpointRemovesDynamicEndpoint(t *testing.T) {
	discovery := &EndpointDiscovery{cachePath: filepath.Join(t.TempDir(), "graphql_ops.json")}
	cache := &EndpointCache{
		Endpoints:   map[string]string{"Followers": "stale-query/Followers"},
		Quarantined: map[string]string{},
		Features:    map[string]bool{},
		OpFeatures:  map[string][]string{},
		Timestamp:   time.Now(),
	}
	discovery.UpdateMemoryCache(cache)
	defer discovery.InvalidateCache()

	manager := &EndpointManager{discovery: discovery}
	manager.QuarantineEndpoint("Followers", "HTTP 404")

	if _, ok := discovery.GetMemoryCache().GetEndpoint("Followers"); ok {
		t.Fatal("quarantined endpoint is still active")
	}
	if !manager.IsQuarantined("Followers") {
		t.Fatal("Followers should be quarantined")
	}
	if got := manager.GetEndpoint("Followers"); got != "Followers" {
		t.Fatalf("quarantined endpoint fallback = %q, want operation name", got)
	}
	if _, ok := manager.ListEndpoints()["Followers"]; ok {
		t.Fatal("quarantined endpoint should not be listed")
	}
}

func TestFollowersHasNoStaticFallback(t *testing.T) {
	if _, ok := GraphQLEndpoints["Followers"]; ok {
		t.Fatal("Followers must not use an unverified static endpoint")
	}
	manager := &EndpointManager{}
	if _, status := manager.CheckEndpoint("Followers"); status != "No verified endpoint" {
		t.Fatalf("Followers status = %q, want no verified endpoint", status)
	}
}
