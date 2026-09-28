package core

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
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

func TestListEndpointsTakesOneCacheSnapshot(t *testing.T) {
	repository := &countingEndpointRepository{cache: &EndpointCache{
		Endpoints:   map[string]string{"DynamicOperation": "dynamic/DynamicOperation"},
		Quarantined: map[string]string{},
		Features:    map[string]bool{},
		OpFeatures:  map[string][]string{},
		Timestamp:   time.Now(),
	}}
	manager := &EndpointManager{discovery: &EndpointDiscovery{}, repository: repository}

	endpoints := manager.ListEndpoints()
	if _, ok := endpoints["DynamicOperation"]; !ok {
		t.Fatal("ListEndpoints() omitted dynamic endpoint")
	}
	if repository.snapshots != 1 {
		t.Fatalf("cache snapshots = %d, want exactly one", repository.snapshots)
	}
}

func TestResolveOperationUsesOneCacheSnapshot(t *testing.T) {
	repository := &countingEndpointRepository{cache: &EndpointCache{
		Endpoints:   map[string]string{"SearchTimeline": "dynamic/SearchTimeline"},
		Quarantined: map[string]string{},
		Features:    map[string]bool{"switch": false},
		OpFeatures:  map[string][]string{"SearchTimeline": {"switch"}},
		Timestamp:   time.Now(),
	}}
	manager := &EndpointManager{discovery: &EndpointDiscovery{}, repository: repository}

	endpoint, features, quarantined := manager.resolveOperation("SearchTimeline")
	if endpoint != "dynamic/SearchTimeline" || features["switch"] || quarantined {
		t.Fatalf("resolved endpoint/features/quarantine = %q/%#v/%t", endpoint, features, quarantined)
	}
	if repository.snapshots != 1 {
		t.Fatalf("cache snapshots = %d, want exactly one", repository.snapshots)
	}
}

func TestPublishDiscoveredEndpointsUpdatesRepositorySnapshot(t *testing.T) {
	repository := &countingEndpointRepository{cache: &EndpointCache{
		Endpoints: map[string]string{"SearchTimeline": "old/SearchTimeline"},
		Timestamp: time.Now(),
	}}
	manager := &EndpointManager{discovery: &EndpointDiscovery{}, repository: repository}
	cache := &EndpointCache{
		Endpoints: map[string]string{"SearchTimeline": "new/SearchTimeline"},
		Timestamp: time.Now(),
	}

	if err := manager.publishDiscoveredEndpoints(cache); err != nil {
		t.Fatalf("publishDiscoveredEndpoints() error = %v", err)
	}
	if got := repository.Snapshot().Endpoints["SearchTimeline"]; got != "new/SearchTimeline" {
		t.Fatalf("repository endpoint = %q, want newly discovered endpoint", got)
	}
}

type countingEndpointRepository struct {
	cache     *EndpointCache
	snapshots int
}

func (r *countingEndpointRepository) Snapshot() *EndpointCache {
	r.snapshots++
	return cloneEndpointCache(r.cache)
}

func (r *countingEndpointRepository) Replace(cache *EndpointCache) error {
	r.cache = cloneEndpointCache(cache)
	return nil
}

func (r *countingEndpointRepository) Mutate(mutator func(*EndpointCache) error) error {
	return mutator(r.cache)
}

func (r *countingEndpointRepository) Invalidate() error {
	r.cache = newEmptyEndpointCache()
	return nil
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

func TestRefreshForOperationHonorsWaitingCallerCancellation(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	manager := &EndpointManager{
		refreshFunc: func(context.Context) error {
			close(started)
			<-release
			return nil
		},
	}

	ownerDone := make(chan error, 1)
	go func() { ownerDone <- manager.RefreshForOperation(context.Background(), "SearchTimeline") }()
	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := manager.RefreshForOperation(ctx, "SearchTimeline"); !errors.Is(err, context.Canceled) {
		t.Fatalf("waiting refresh error = %v, want context cancellation", err)
	}
	close(release)
	if err := <-ownerDone; err != nil {
		t.Fatalf("owner refresh error = %v", err)
	}
}

func TestRefreshForOperationCoalescesConcurrentCallers(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var refreshes atomic.Int32
	var startedOnce sync.Once
	manager := &EndpointManager{
		refreshFunc: func(context.Context) error {
			refreshes.Add(1)
			startedOnce.Do(func() { close(started) })
			<-release
			return nil
		},
	}

	const callers = 8
	results := make(chan error, callers)
	for i := 0; i < callers; i++ {
		go func() { results <- manager.RefreshForOperation(context.Background(), "SearchTimeline") }()
	}
	<-started
	time.Sleep(20 * time.Millisecond)
	close(release)
	for i := 0; i < callers; i++ {
		if err := <-results; err != nil {
			t.Fatalf("coalesced refresh error = %v", err)
		}
	}
	if got := refreshes.Load(); got != 1 {
		t.Fatalf("refreshes = %d, want 1", got)
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
