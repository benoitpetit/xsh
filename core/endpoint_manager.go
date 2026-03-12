// Package core provides dynamic endpoint management with auto-discovery
// This file now delegates all operations to EndpointDiscovery for a unified cache system.
package core

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// EndpointManager manages GraphQL endpoints with automatic discovery and updates
// Now uses EndpointDiscovery as the single source of truth for caching.
type EndpointManager struct {
	discovery *EndpointDiscovery
	verbose   bool
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
			discovery: discovery,
			verbose:   Verbose,
		}
	})

	return globalEndpointManager
}

// getCache returns the current cache from discovery
func (em *EndpointManager) getCache() *EndpointCache {
	if em.discovery == nil {
		return nil
	}
	
	// Try to get from memory cache first
	if cache := em.discovery.GetMemoryCache(); cache != nil && cache.IsValid() {
		return cache
	}
	
	// Try to load from disk
	cache, err := em.discovery.LoadCache()
	if err == nil && cache.IsValid() {
		em.discovery.UpdateMemoryCache(cache)
		return cache
	}
	
	// Return empty cache that will trigger fallback
	return &EndpointCache{
		Endpoints:  make(map[string]string),
		Features:   make(map[string]bool),
		OpFeatures: make(map[string][]string),
	}
}

// GetEndpoint returns the endpoint for an operation, with auto-discovery
func (em *EndpointManager) GetEndpoint(operation string) string {
	cache := em.getCache()
	
	// Check dynamic cache first
	if cache != nil {
		if endpoint, ok := cache.Endpoints[operation]; ok {
			return endpoint
		}
	}
	
	// Check static fallback
	if endpoint, ok := GraphQLEndpoints[operation]; ok {
		if em.verbose {
			log.Printf("[EndpointManager] Using static fallback for %s", operation)
		}
		return endpoint
	}
	
	return operation
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
			
			if _, err := em.discovery.DiscoverEndpoints(ctx); err != nil {
				// Return existing endpoint even if refresh fails
				if em.verbose {
					log.Printf("[EndpointManager] Refresh failed, using cached: %v", err)
				}
			}
		}
	}
	
	return endpoint, nil
}

// GetOpFeatures returns feature switches for a specific operation
func (em *EndpointManager) GetOpFeatures(operation string) map[string]bool {
	result := make(map[string]bool)
	
	cache := em.getCache()
	if cache == nil {
		// Return default features
		for k, v := range DefaultFeatures {
			result[k] = v
		}
		return result
	}
	
	keys, ok := cache.OpFeatures[operation]
	if !ok || len(keys) == 0 {
		// Return default features
		for k, v := range DefaultFeatures {
			result[k] = v
		}
		return result
	}
	
	// Get specific features for this operation
	for _, key := range keys {
		if val, ok := cache.Features[key]; ok {
			result[key] = val
		} else {
			result[key] = true // Default to true
		}
	}
	
	return result
}

// RefreshEndpoints fetches fresh endpoints from X.com
func (em *EndpointManager) RefreshEndpoints(ctx context.Context) error {
	if em.discovery == nil {
		return fmt.Errorf("discovery not available")
	}
	
	_, err := em.discovery.DiscoverEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}
	
	if em.verbose {
		cache := em.getCache()
		if cache != nil {
			log.Printf("[EndpointManager] Refreshed %d endpoints", len(cache.Endpoints))
		}
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
	
	// Update in discovery cache
	cache := em.getCache()
	if cache == nil {
		cache = &EndpointCache{
			Endpoints:  make(map[string]string),
			Features:   make(map[string]bool),
			OpFeatures: make(map[string][]string),
			Timestamp:  time.Now(),
		}
	}
	
	cache.Endpoints[operation] = endpoint
	cache.Timestamp = time.Now()
	
	// Save to disk
	if err := em.discovery.SaveCache(cache); err != nil {
		log.Printf("[EndpointManager] Warning: failed to save cache: %v", err)
	}
	
	em.discovery.UpdateMemoryCache(cache)
	
	if em.verbose {
		log.Printf("[EndpointManager] Updated endpoint %s -> %s", operation, endpoint)
	}
}

// ResetEndpoint resets an endpoint to its default value
func (em *EndpointManager) ResetEndpoint(operation string) {
	if em.discovery == nil {
		return
	}
	
	cache := em.getCache()
	if cache == nil {
		return
	}
	
	delete(cache.Endpoints, operation)
	cache.Timestamp = time.Now()
	
	// Save to disk
	if err := em.discovery.SaveCache(cache); err != nil {
		log.Printf("[EndpointManager] Warning: failed to save cache: %v", err)
	}
	
	em.discovery.UpdateMemoryCache(cache)
}

// ListEndpoints returns all current endpoints
func (em *EndpointManager) ListEndpoints() map[string]string {
	result := make(map[string]string)
	
	// Start with static fallbacks
	for k, v := range GraphQLEndpoints {
		result[k] = v
	}
	
	// Override with dynamic endpoints
	cache := em.getCache()
	if cache != nil {
		for k, v := range cache.Endpoints {
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
			TotalCount:   len(GraphQLEndpoints),
			FeatureCount: len(DefaultFeatures),
			LastUpdated:  time.Time{},
			CacheAge:     0,
		}
	}
	
	return EndpointStats{
		TotalCount:   len(cache.Endpoints),
		FeatureCount: len(cache.Features),
		LastUpdated:  cache.Timestamp,
		CacheAge:     time.Since(cache.Timestamp),
	}
}

// EndpointStats represents endpoint statistics
type EndpointStats struct {
	TotalCount   int           `json:"total_count"`
	FeatureCount int           `json:"feature_count"`
	LastUpdated  time.Time     `json:"last_updated"`
	CacheAge     time.Duration `json:"cache_age"`
}

// CheckEndpoint checks if an endpoint is valid
func (em *EndpointManager) CheckEndpoint(operation string) (bool, string) {
	cache := em.getCache()
	
	// Check if using dynamic endpoint
	isDynamic := false
	if cache != nil {
		_, isDynamic = cache.Endpoints[operation]
	}
	
	endpoint := em.GetEndpoint(operation)
	
	if !isDynamic {
		return false, fmt.Sprintf("Using static fallback: %s", endpoint)
	}
	
	return true, fmt.Sprintf("Using dynamic: %s", endpoint)
}

// Invalidate clears all dynamic endpoints
func (em *EndpointManager) Invalidate() {
	if em.discovery != nil {
		em.discovery.InvalidateCache()
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
		"not found",
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
