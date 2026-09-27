package core

import (
	"fmt"
	"sync"
)

// EndpointRepository serializes endpoint cache updates and exposes immutable
// snapshots to callers.
type EndpointRepository interface {
	Snapshot() *EndpointCache
	Replace(*EndpointCache) error
	Mutate(func(*EndpointCache) error) error
	Invalidate() error
}

type endpointRepository struct {
	discovery *EndpointDiscovery
	mu        sync.RWMutex
	cache     *EndpointCache
}

// NewEndpointRepository creates a repository backed by discovery's durable
// cache when a discovery instance is supplied.
func NewEndpointRepository(discovery *EndpointDiscovery) EndpointRepository {
	repository := &endpointRepository{discovery: discovery}
	if discovery != nil {
		if cache := discovery.GetMemoryCache(); cache != nil {
			repository.cache = cloneEndpointCache(cache)
		} else if cache, err := discovery.LoadCache(); err == nil {
			repository.cache = cloneEndpointCache(cache)
		}
	}
	return repository
}

func (r *endpointRepository) Snapshot() *EndpointCache {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return cloneEndpointCache(r.cache)
}

func (r *endpointRepository) Replace(cache *EndpointCache) error {
	if cache == nil {
		return fmt.Errorf("endpoint cache must not be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	next := cloneEndpointCache(cache)
	mergeQuarantines(r.cache, next)
	if err := r.persistLocked(next); err != nil {
		return err
	}
	r.cache = next
	return nil
}

func (r *endpointRepository) Mutate(mutator func(*EndpointCache) error) error {
	if mutator == nil {
		return fmt.Errorf("endpoint cache mutator must not be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	next := cloneEndpointCache(r.cache)
	if next == nil {
		next = newEmptyEndpointCache()
	}
	if err := mutator(next); err != nil {
		return err
	}
	normalizeEndpointCache(next)
	if err := r.persistLocked(next); err != nil {
		return err
	}
	r.cache = next
	return nil
}

func (r *endpointRepository) Invalidate() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = newEmptyEndpointCache()
	if r.discovery != nil {
		r.discovery.InvalidateCache()
	}
	return nil
}

func (r *endpointRepository) persistLocked(cache *EndpointCache) error {
	if r.discovery == nil {
		return nil
	}
	if err := r.discovery.SaveCache(cache); err != nil {
		return err
	}
	r.discovery.UpdateMemoryCache(cache)
	return nil
}

func mergeQuarantines(previous, next *EndpointCache) {
	if previous == nil || len(previous.Quarantined) == 0 {
		return
	}
	if next.Quarantined == nil {
		next.Quarantined = make(map[string]string)
	}
	for operation, reason := range previous.Quarantined {
		if _, replaced := next.Endpoints[operation]; replaced {
			delete(next.Quarantined, operation)
			continue
		}
		if _, explicitlyQuarantined := next.Quarantined[operation]; !explicitlyQuarantined {
			next.Quarantined[operation] = reason
		}
		delete(next.Endpoints, operation)
	}
}

func newEmptyEndpointCache() *EndpointCache {
	return &EndpointCache{
		Endpoints:   make(map[string]string),
		Quarantined: make(map[string]string),
		Features:    make(map[string]bool),
		OpFeatures:  make(map[string][]string),
	}
}

func normalizeEndpointCache(cache *EndpointCache) {
	if cache.Endpoints == nil {
		cache.Endpoints = make(map[string]string)
	}
	if cache.Quarantined == nil {
		cache.Quarantined = make(map[string]string)
	}
	if cache.Features == nil {
		cache.Features = make(map[string]bool)
	}
	if cache.OpFeatures == nil {
		cache.OpFeatures = make(map[string][]string)
	}
}

func cloneEndpointCache(cache *EndpointCache) *EndpointCache {
	if cache == nil {
		return nil
	}
	copy := *cache
	copy.Endpoints = make(map[string]string, len(cache.Endpoints))
	for key, value := range cache.Endpoints {
		copy.Endpoints[key] = value
	}
	copy.Quarantined = make(map[string]string, len(cache.Quarantined))
	for key, value := range cache.Quarantined {
		copy.Quarantined[key] = value
	}
	copy.Features = make(map[string]bool, len(cache.Features))
	for key, value := range cache.Features {
		copy.Features[key] = value
	}
	copy.OpFeatures = make(map[string][]string, len(cache.OpFeatures))
	for key, values := range cache.OpFeatures {
		copy.OpFeatures[key] = append([]string(nil), values...)
	}
	return &copy
}
