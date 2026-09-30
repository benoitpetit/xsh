// Package core provides dynamic endpoint management with auto-discovery
// This file now delegates all operations to EndpointDiscovery for a unified cache system.
package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// EndpointManager manages GraphQL endpoints with automatic discovery and updates
// Now uses EndpointDiscovery as the single source of truth for caching.
type EndpointManager struct {
	discovery       *EndpointDiscovery
	repository      EndpointRepository
	repositoryMu    sync.Mutex
	verbose         bool
	refreshCooldown time.Duration
	refreshFunc     func(context.Context) error

	refreshMu    sync.Mutex
	refreshState *endpointRefreshState
	failedAt     map[string]time.Time
}

const defaultEndpointRefreshCooldown = 30 * time.Second

type endpointRefreshState struct {
	done chan struct{}
	err  error
}

// Global endpoint manager instance
var (
	globalEndpointManager     *EndpointManager
	globalEndpointManagerOnce sync.Once
)

// GetEndpointManager returns the singleton endpoint manager
func GetEndpointManager() *EndpointManager {
	globalEndpointManagerOnce.Do(func() {
		discovery, err := NewEndpointDiscovery(Verbose)
		if err != nil {
			log.Printf("[EndpointManager] Warning: failed to create discovery: %v", err)
		}

		globalEndpointManager = &EndpointManager{
			discovery:  discovery,
			repository: NewEndpointRepository(discovery),
			verbose:    Verbose,
		}
	})

	return globalEndpointManager
}

// getCache returns the current cache from discovery
func (em *EndpointManager) getCache() *EndpointCache {
	repository := em.getRepository()
	if repository == nil {
		return nil
	}
	cache := repository.Snapshot()
	if em == globalEndpointManager {
		if latest := GetMemoryCache(); latest != nil && latest.IsUsable() && (cache == nil || latest.Timestamp.After(cache.Timestamp)) {
			if err := repository.Replace(latest); err == nil {
				cache = latest
			}
		}
	}
	if cache != nil && cache.IsUsable() {
		return cache
	}

	cache, err := em.discovery.LoadCache()
	if err == nil && cache.IsUsable() {
		_ = repository.Replace(cache)
		return cache
	}

	// Return empty cache that will trigger fallback
	return &EndpointCache{
		Endpoints:      make(map[string]string),
		Quarantined:    make(map[string]string),
		QuarantinedIDs: make(map[string]string),
		Features:       make(map[string]bool),
		OpFeatures:     make(map[string][]string),
	}
}

func (em *EndpointManager) getRepository() EndpointRepository {
	if em == nil || em.discovery == nil {
		return nil
	}
	em.repositoryMu.Lock()
	defer em.repositoryMu.Unlock()
	if em.repository == nil {
		em.repository = NewEndpointRepository(em.discovery)
	}
	return em.repository
}

// GetEndpoint returns the endpoint for an operation, with auto-discovery
func (em *EndpointManager) GetEndpoint(operation string) string {
	return em.resolveEndpoint(em.getCache(), operation)
}

// resolveOperation reads endpoint, feature flags, and quarantine state from a
// single immutable cache snapshot. Request paths should use this instead of
// separately asking for each piece of operation metadata.
func (em *EndpointManager) resolveOperation(operation string) (string, map[string]bool, bool) {
	cache := em.getCache()
	return em.resolveEndpoint(cache, operation), operationFeatures(cache, operation), isOperationQuarantined(cache, operation)
}

// resolveOperationWithRefresh gives quarantined operations a chance to recover
// when the endpoint cache is old. Recent quarantines remain blocked so every
// command does not trigger a discovery request for the same confirmed 404.
func (em *EndpointManager) resolveOperationWithRefresh(ctx context.Context, operation string) (string, map[string]bool, bool, error) {
	endpoint, features, quarantined := em.resolveOperation(operation)
	if !quarantined {
		return endpoint, features, false, nil
	}

	cache := em.getCache()
	if cache == nil || !cache.IsStale() {
		return endpoint, features, true, nil
	}

	if err := em.RefreshForOperation(ctx, operation); err != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
		return endpoint, features, true, err
	}
	endpoint, features, quarantined = em.resolveOperation(operation)
	return endpoint, features, quarantined, nil
}

func (em *EndpointManager) resolveEndpoint(cache *EndpointCache, operation string) string {
	if isOperationQuarantined(cache, operation) {
		return operation
	}
	if cache != nil {
		if endpoint, ok := cache.Endpoints[operation]; ok {
			return endpoint
		}
	}
	if endpoint, ok := GraphQLEndpoints[operation]; ok {
		if em.verbose {
			log.Printf("[EndpointManager] Using static fallback for %s", operation)
		}
		return endpoint
	}
	return operation
}

func isOperationQuarantined(cache *EndpointCache, operation string) bool {
	if cache == nil {
		return false
	}
	_, quarantined := cache.Quarantined[operation]
	return quarantined
}

func operationFeatures(cache *EndpointCache, operation string) map[string]bool {
	features := make(map[string]bool)
	if cache == nil {
		for key, value := range DefaultFeatures {
			features[key] = value
		}
	} else if keys := cache.OpFeatures[operation]; len(keys) == 0 {
		for key, value := range DefaultFeatures {
			features[key] = value
		}
	} else {
		for _, key := range cache.OpFeatures[operation] {
			if value, ok := cache.Features[key]; ok {
				features[key] = value
			} else {
				features[key] = true
			}
		}
	}
	return features
}

// GetEndpointWithRefresh returns endpoint, refreshing if necessary
func (em *EndpointManager) GetEndpointWithRefresh(ctx context.Context, operation string) (string, error) {
	endpoint := em.GetEndpoint(operation)

	// Check if we need to refresh
	cache := em.getCache()
	if cache == nil || !cache.IsValid() {
		if em.discovery != nil {
			if em.verbose {
				log.Println("[EndpointManager] Cache expired, refreshing endpoints...")
			}

			if err := em.RefreshForOperation(ctx, operation); err != nil {
				// Return existing endpoint even if refresh fails
				if em.verbose {
					log.Printf("[EndpointManager] Refresh failed, using cached: %v", err)
				}
			} else {
				// Resolve again so this request uses the freshly discovered ID,
				// rather than the fallback selected before the refresh.
				endpoint = em.GetEndpoint(operation)
			}
		}
	}

	return endpoint, nil
}

// RefreshForOperation refreshes endpoint discovery once per operation and
// coalesces concurrent callers. Failed operations are cooled down briefly so
// a removed GraphQL operation cannot trigger a discovery storm.
func (em *EndpointManager) RefreshForOperation(ctx context.Context, operation string) error {
	em.refreshMu.Lock()
	cooldown := em.refreshCooldown
	if cooldown <= 0 {
		cooldown = defaultEndpointRefreshCooldown
	}
	if em.failedAt != nil {
		if failedAt, ok := em.failedAt[operation]; ok && time.Since(failedAt) < cooldown {
			err := fmt.Errorf("endpoint refresh for %q is cooling down after a failure", operation)
			em.refreshMu.Unlock()
			return err
		}
	}
	if state := em.refreshState; state != nil {
		em.refreshMu.Unlock()
		select {
		case <-state.done:
			return state.err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	state := &endpointRefreshState{done: make(chan struct{})}
	em.refreshState = state
	em.refreshMu.Unlock()

	if em.refreshFunc != nil {
		state.err = em.refreshFunc(ctx)
	} else {
		state.err = em.RefreshEndpoints(ctx)
	}

	em.refreshMu.Lock()
	if state.err != nil {
		if em.failedAt == nil {
			em.failedAt = make(map[string]time.Time)
		}
		em.failedAt[operation] = time.Now()
	} else if em.failedAt != nil {
		delete(em.failedAt, operation)
	}
	em.refreshState = nil
	close(state.done)
	em.refreshMu.Unlock()
	return state.err
}

// GetOpFeatures returns feature switches for a specific operation
func (em *EndpointManager) GetOpFeatures(operation string) map[string]bool {
	return operationFeatures(em.getCache(), operation)
}

// GetOpFeaturesWithSource returns one consistent feature snapshot and identifies
// whether it came from operation metadata or the built-in fallback.
func (em *EndpointManager) GetOpFeaturesWithSource(operation string) (map[string]bool, string) {
	cache := em.getCache()
	if cache != nil && len(cache.OpFeatures[operation]) > 0 {
		return operationFeatures(cache, operation), "operation_discovery"
	}
	return operationFeatures(cache, operation), "built_in_defaults"
}

// RefreshEndpoints fetches fresh endpoints from X.com
func (em *EndpointManager) RefreshEndpoints(ctx context.Context) error {
	return em.RefreshEndpointsForAccount(ctx, "")
}

// RefreshEndpointsForAccount discovers IDs with the explicitly selected session.
func (em *EndpointManager) RefreshEndpointsForAccount(ctx context.Context, account string) error {
	if em.discovery == nil {
		return fmt.Errorf("discovery not available")
	}

	discovery := em.discovery
	if account != "" {
		var err error
		discovery, err = NewEndpointDiscoveryForAccount(em.verbose, account)
		if err != nil {
			return err
		}
	}
	cache, err := discovery.DiscoverEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}
	if err := em.publishDiscoveredEndpoints(cache); err != nil {
		return fmt.Errorf("failed to publish discovered endpoints: %w", err)
	}

	if em.verbose {
		cache := em.getCache()
		if cache != nil {
			log.Printf("[EndpointManager] Refreshed %d endpoints", len(cache.Endpoints))
		}
	}

	return nil
}

func (em *EndpointManager) publishDiscoveredEndpoints(cache *EndpointCache) error {
	if repository := em.getRepository(); repository != nil {
		return repository.Replace(cache)
	}
	return nil
}

// CheckAndUpdate checks if endpoints need updating and updates if necessary
func (em *EndpointManager) CheckAndUpdate(ctx context.Context) error {
	cache := em.getCache()

	// Check if update needed (cache expired)
	if cache == nil || !cache.IsValid() {
		return em.RefreshEndpoints(ctx)
	}

	return nil
}

// UpdateEndpoint manually updates a single endpoint
func (em *EndpointManager) UpdateEndpoint(operation, endpoint string) {
	if em.discovery == nil {
		return
	}
	if err := em.getRepository().Mutate(func(cache *EndpointCache) error {
		cache.Endpoints[operation] = endpoint
		delete(cache.Quarantined, operation)
		delete(cache.QuarantinedIDs, operation)
		cache.Timestamp = time.Now()
		return nil
	}); err != nil {
		log.Printf("[EndpointManager] Warning: failed to save cache: %v", err)
	}

	if em.verbose {
		log.Printf("[EndpointManager] Updated endpoint %s -> %s", operation, endpoint)
	}
}

// ResetEndpoint resets an endpoint to its default value
func (em *EndpointManager) ResetEndpoint(operation string) {
	if em.discovery == nil {
		return
	}

	if err := em.getRepository().Mutate(func(cache *EndpointCache) error {
		delete(cache.Endpoints, operation)
		delete(cache.Quarantined, operation)
		delete(cache.QuarantinedIDs, operation)
		cache.Timestamp = time.Now()
		return nil
	}); err != nil {
		log.Printf("[EndpointManager] Warning: failed to save cache: %v", err)
	}
}

// QuarantineEndpoint removes an operation from active endpoint selection after
// X returns a definitive not-found response. The quarantine is persisted with
// the cache so repeated calls do not keep retrying the same stale endpoint.
func (em *EndpointManager) QuarantineEndpoint(operation, reason string) {
	if em.discovery == nil || operation == "" {
		return
	}
	if reason == "" {
		reason = "endpoint returned not found"
	}
	if err := em.getRepository().Mutate(func(cache *EndpointCache) error {
		rejectedID := cache.Endpoints[operation]
		if rejectedID == "" {
			rejectedID = cache.QuarantinedIDs[operation]
		}
		if rejectedID == "" {
			rejectedID = GraphQLEndpoints[operation]
		}
		delete(cache.Endpoints, operation)
		cache.Quarantined[operation] = reason
		if rejectedID != "" {
			cache.QuarantinedIDs[operation] = rejectedID
		}
		cache.Timestamp = time.Now()
		return nil
	}); err != nil {
		log.Printf("[EndpointManager] Warning: failed to save quarantined endpoint: %v", err)
	}
}

// RecordProbeResult persists a confirmed obsolete operation so later requests
// do not keep selecting the same endpoint. Other probe outcomes are read-only.
func (em *EndpointManager) RecordProbeResult(operation string, probe EndpointProbe) bool {
	if probe.State != "obsolete" {
		return false
	}
	em.QuarantineEndpoint(operation, probe.Message)
	return em.IsQuarantined(operation)
}

// IsQuarantined reports whether an operation has been isolated after a 404.
func (em *EndpointManager) IsQuarantined(operation string) bool {
	cache := em.getCache()
	if cache == nil {
		return false
	}
	_, ok := cache.Quarantined[operation]
	return ok
}

// IsRejectedEndpoint reports whether discovery found the same ID that was
// previously confirmed obsolete. Previews use it to mirror actual publishing.
func (em *EndpointManager) IsRejectedEndpoint(operation, endpoint string) bool {
	cache := em.getCache()
	return cache != nil && cache.QuarantinedIDs[operation] != "" && cache.QuarantinedIDs[operation] == endpoint
}

// ListEndpoints returns all current endpoints
func (em *EndpointManager) ListEndpoints() map[string]string {
	result := make(map[string]string)
	cache := em.getCache()

	// Start with static fallbacks
	for k, v := range GraphQLEndpoints {
		if cache != nil {
			if _, quarantined := cache.Quarantined[k]; quarantined {
				continue
			}
		}
		result[k] = v
	}

	// Override with dynamic endpoints
	if cache != nil {
		for k, v := range cache.Endpoints {
			if _, quarantined := cache.Quarantined[k]; quarantined {
				continue
			}
			result[k] = v
		}
	}

	return result
}

// GetStats returns statistics about endpoints
func (em *EndpointManager) GetStats() EndpointStats {
	cache := em.getCache()

	if cache == nil {
		return EndpointStats{
			TotalCount:       len(GraphQLEndpoints),
			StaticCount:      len(GraphQLEndpoints),
			DynamicCount:     0,
			QuarantinedCount: 0,
			FeatureCount:     len(DefaultFeatures),
			LastUpdated:      time.Time{},
			CacheAge:         0,
		}
	}

	staticCount := 0
	for operation := range GraphQLEndpoints {
		if _, quarantined := cache.Quarantined[operation]; quarantined {
			continue
		}
		if _, ok := cache.Endpoints[operation]; !ok {
			staticCount++
		}
	}

	cacheAge := time.Duration(0)
	if !cache.Timestamp.IsZero() {
		cacheAge = time.Since(cache.Timestamp)
	}

	return EndpointStats{
		TotalCount:       staticCount + len(cache.Endpoints),
		StaticCount:      staticCount,
		DynamicCount:     len(cache.Endpoints),
		QuarantinedCount: len(cache.Quarantined),
		FeatureCount:     len(cache.Features),
		LastUpdated:      cache.Timestamp,
		CacheAge:         cacheAge,
	}
}

// EndpointStats represents endpoint statistics
type EndpointStats struct {
	TotalCount       int           `json:"total_count"`
	StaticCount      int           `json:"static_count"`
	DynamicCount     int           `json:"dynamic_count"`
	QuarantinedCount int           `json:"quarantined_count"`
	FeatureCount     int           `json:"feature_count"`
	LastUpdated      time.Time     `json:"last_updated"`
	CacheAge         time.Duration `json:"cache_age"`
}

// CheckEndpoint checks if an endpoint is valid
func (em *EndpointManager) CheckEndpoint(operation string) (bool, string) {
	cache := em.getCache()

	// Check if using dynamic endpoint
	isDynamic := false
	if cache != nil {
		if _, quarantined := cache.Quarantined[operation]; quarantined {
			return false, "Quarantined after 404"
		}
		_, isDynamic = cache.Endpoints[operation]
	}

	endpoint := em.GetEndpoint(operation)

	if !isDynamic {
		if _, hasStaticFallback := GraphQLEndpoints[operation]; !hasStaticFallback {
			return false, "No verified endpoint"
		}
		return false, fmt.Sprintf("Using static fallback: %s", endpoint)
	}

	return true, fmt.Sprintf("Using dynamic: %s", endpoint)
}

// Invalidate clears all dynamic endpoints
func (em *EndpointManager) Invalidate() {
	if em.discovery != nil {
		_ = em.getRepository().Invalidate()
	}
}

// Global functions for easy access

// GetGraphQLEndpoints returns the current endpoint map
func GetGraphQLEndpoints() map[string]string {
	return GetEndpointManager().ListEndpoints()
}

// GetEndpoint returns a single endpoint
func GetEndpoint(operation string) string {
	return GetEndpointManager().GetEndpoint(operation)
}

// RefreshEndpointsGlobal triggers a refresh of all endpoints
func RefreshEndpointsGlobal() error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	return GetEndpointManager().RefreshEndpoints(ctx)
}

// InvalidateCache clears all cached endpoints
func InvalidateCache() {
	GetEndpointManager().Invalidate()
	if discovery, err := NewEndpointDiscovery(Verbose); err == nil {
		discovery.InvalidateCache()
	}
}

// IsEndpointObsolete checks if an error indicates an obsolete endpoint
func IsEndpointObsolete(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	indicators := []string{
		"query not found",
		"http 404",
		"notfounderror",
		"resource not found",
		"graphql endpoint",
		"operation not found",
	}

	for _, indicator := range indicators {
		if strings.Contains(errStr, indicator) {
			return true
		}
	}

	return false
}

// GetEndpointSuggestion returns a helpful message for updating endpoints
func GetEndpointSuggestion(operation string) string {
	return fmt.Sprintf(`
The GraphQL endpoint for '%s' appears to be obsolete (404 error).

To fix this, try refreshing endpoints from X.com:

  xsh endpoints refresh

Or check the current endpoint status:

  xsh endpoints check %s

  xsh endpoints list
`, operation, operation)
}
