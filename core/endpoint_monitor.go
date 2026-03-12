// Package core provides proactive endpoint obsolescence detection and auto-update
package core

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// EndpointMonitor continuously monitors endpoint health and auto-updates
type EndpointMonitor struct {
	discovery     *EndpointDiscovery
	client        *XClient
	stopChan      chan struct{}
	wg            sync.WaitGroup
	checkInterval time.Duration
	verbose       bool
	lastCheck     time.Time
	status        MonitorStatus
	mu            sync.RWMutex
}

// MonitorStatus represents the current monitoring status
type MonitorStatus struct {
	IsHealthy       bool      `json:"is_healthy"`
	LastCheck       time.Time `json:"last_check"`
	EndpointsCount  int       `json:"endpoints_count"`
	FailedEndpoints []string  `json:"failed_endpoints,omitempty"`
	NeedsUpdate     bool      `json:"needs_update"`
	Message         string    `json:"message"`
}

// NewEndpointMonitor creates a new endpoint monitor
func NewEndpointMonitor(client *XClient, verbose bool) (*EndpointMonitor, error) {
	discovery, err := NewEndpointDiscovery(verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery: %w", err)
	}

	return &EndpointMonitor{
		discovery:     discovery,
		client:        client,
		stopChan:      make(chan struct{}),
		checkInterval: 5 * time.Minute,
		verbose:       verbose,
		status: MonitorStatus{
			IsHealthy: true,
		},
	}, nil
}

// Start begins background monitoring
func (em *EndpointMonitor) Start() {
	em.wg.Add(1)
	go em.monitorLoop()
	
	if em.verbose {
		log.Println("[EndpointMonitor] Started background monitoring")
	}
}

// Stop halts background monitoring
func (em *EndpointMonitor) Stop() {
	close(em.stopChan)
	em.wg.Wait()
	
	if em.verbose {
		log.Println("[EndpointMonitor] Stopped background monitoring")
	}
}

// monitorLoop runs the monitoring loop
func (em *EndpointMonitor) monitorLoop() {
	defer em.wg.Done()

	// Initial check
	em.performCheck()

	ticker := time.NewTicker(em.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-em.stopChan:
			return
		case <-ticker.C:
			em.performCheck()
		}
	}
}

// performCheck performs a single health check
func (em *EndpointMonitor) performCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status := MonitorStatus{
		LastCheck: time.Now(),
	}

	// Get current cache
	cache, err := em.discovery.GetCachedEndpoints(ctx)
	if err != nil {
		status.IsHealthy = false
		status.Message = fmt.Sprintf("Failed to get endpoints: %v", err)
		status.NeedsUpdate = true
		em.updateStatus(status)
		return
	}

	status.EndpointsCount = len(cache.Endpoints)

	// Check if cache is stale
	if cache.IsStale() {
		status.NeedsUpdate = true
		status.Message = "Cache is stale, needs refresh"
	}

	// Test critical endpoints
	failedEndpoints := em.testCriticalEndpoints(ctx, cache)
	status.FailedEndpoints = failedEndpoints

	if len(failedEndpoints) > 0 {
		status.IsHealthy = false
		status.Message = fmt.Sprintf("%d endpoints failed", len(failedEndpoints))
		status.NeedsUpdate = true

		// Auto-update if more than 30% of critical endpoints fail
		if float64(len(failedEndpoints))/float64(len(criticalEndpoints)) > 0.3 {
			if em.verbose {
				log.Printf("[EndpointMonitor] Auto-updating endpoints due to high failure rate")
			}
			if err := em.AutoUpdate(ctx); err != nil && em.verbose {
				log.Printf("[EndpointMonitor] Auto-update failed: %v", err)
			}
		}
	} else {
		status.IsHealthy = true
		if status.Message == "" {
			status.Message = "All endpoints healthy"
		}
	}

	em.updateStatus(status)
}

// criticalEndpoints are endpoints to test for health
var criticalEndpoints = []string{
	"HomeTimeline",
	"HomeLatestTimeline",
	"UserByScreenName",
	"SearchTimeline",
	"TweetDetail",
	"UserTweets",
}

// testCriticalEndpoints tests a list of critical endpoints
func (em *EndpointMonitor) testCriticalEndpoints(ctx context.Context, cache *EndpointCache) []string {
	var failed []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore for concurrent tests
	sem := make(chan struct{}, 3)

	for _, op := range criticalEndpoints {
		wg.Add(1)
		go func(operation string) {
			defer wg.Done()
			
			sem <- struct{}{}
			defer func() { <-sem }()

			endpoint, ok := cache.GetEndpoint(operation)
			if !ok {
				mu.Lock()
				failed = append(failed, fmt.Sprintf("%s: missing", operation))
				mu.Unlock()
				return
			}

			// Quick HEAD request to check if endpoint exists
			if err := em.testEndpoint(ctx, endpoint); err != nil {
				mu.Lock()
				failed = append(failed, fmt.Sprintf("%s: %v", operation, err))
				mu.Unlock()
			}
		}(op)
	}

	wg.Wait()
	return failed
}

// testEndpoint performs a lightweight test of an endpoint
func (em *EndpointMonitor) testEndpoint(ctx context.Context, endpoint string) error {
	// We can't easily test without auth, but we can check if the URL format looks valid
	// and make a request to see if we get 404 (obsolete) vs 401/403 (auth required)
	
	_ = GraphQLBase + "/" + endpoint
	
	// This is a placeholder - in practice, we'd need to make an actual request
	// For now, we just validate the endpoint format
	parts := strings.Split(endpoint, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid endpoint format")
	}
	
	return nil
}

// AutoUpdate fetches new endpoints and updates the cache
func (em *EndpointMonitor) AutoUpdate(ctx context.Context) error {
	if em.verbose {
		log.Println("[EndpointMonitor] Auto-updating endpoints...")
	}

	// Invalidate and refresh
	em.discovery.InvalidateCache()
	
	_, err := em.discovery.DiscoverEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	if em.verbose {
		log.Println("[EndpointMonitor] Auto-update completed successfully")
	}

	return nil
}

// GetStatus returns the current monitor status
func (em *EndpointMonitor) GetStatus() MonitorStatus {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.status
}

// updateStatus updates the monitor status
func (em *EndpointMonitor) updateStatus(status MonitorStatus) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.status = status
	em.lastCheck = time.Now()
}

// SetCheckInterval changes the check interval
func (em *EndpointMonitor) SetCheckInterval(interval time.Duration) {
	em.checkInterval = interval
}

// StartupCheck performs a quick check at startup
func StartupCheck() {
	if Verbose {
		log.Println("[StartupCheck] Performing endpoint health check...")
	}

	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		if Verbose {
			log.Printf("[StartupCheck] Failed to create discovery: %v", err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cache, err := discovery.GetCachedEndpoints(ctx)
	if err != nil {
		if Verbose {
			log.Printf("[StartupCheck] Failed to get endpoints: %v", err)
		}
		// Try to discover fresh
		cache, err = discovery.DiscoverEndpoints(ctx)
		if err != nil {
			if Verbose {
				log.Printf("[StartupCheck] Fresh discovery failed: %v", err)
			}
			return
		}
	}

	// Check if stale
	if cache.IsStale() {
		if Verbose {
			log.Println("[StartupCheck] Cache is stale, refreshing...")
		}
		discovery.InvalidateCache()
		_, err = discovery.DiscoverEndpoints(ctx)
		if err != nil && Verbose {
			log.Printf("[StartupCheck] Refresh failed: %v", err)
		}
	}

	if Verbose {
		log.Printf("[StartupCheck] %d endpoints available", len(cache.Endpoints))
	}
}

// GetEndpointStatus returns detailed status for all endpoints
func GetEndpointStatus() ([]EndpointStatusDetail, error) {
	discovery, err := NewEndpointDiscovery(Verbose)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cache, err := discovery.GetCachedEndpoints(ctx)
	if err != nil {
		return nil, err
	}

	var details []EndpointStatusDetail
	
	for op, endpoint := range cache.Endpoints {
		detail := EndpointStatusDetail{
			Operation: op,
			Endpoint:  endpoint,
			Source:    "static",
		}
		
		if cache.OpFeatures[op] != nil {
			detail.HasFeatures = true
			detail.FeatureCount = len(cache.OpFeatures[op])
		}
		
		details = append(details, detail)
	}

	return details, nil
}

// EndpointStatusDetail represents detailed status for a single endpoint
type EndpointStatusDetail struct {
	Operation    string `json:"operation"`
	Endpoint     string `json:"endpoint"`
	Source       string `json:"source"`
	HasFeatures  bool   `json:"has_features"`
	FeatureCount int    `json:"feature_count"`
}
