// Package core provides intelligent caching for API responses
package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ResponseCache provides intelligent caching for API responses
type ResponseCache struct {
	mu       sync.RWMutex
	entries  map[string]*CacheEntry
	maxSize  int
	maxAge   time.Duration
	cacheDir string
}

// CacheEntry represents a cached response
type CacheEntry struct {
	Data      interface{}   `json:"data"`
	Timestamp time.Time     `json:"timestamp"`
	Key       string        `json:"key"`
	TTL       time.Duration `json:"ttl"`
}

// IsValid checks if cache entry is still valid
func (e *CacheEntry) IsValid() bool {
	return time.Since(e.Timestamp) < e.TTL
}

// NewResponseCache creates a new response cache
func NewResponseCache(maxSize int, maxAge time.Duration) (*ResponseCache, error) {
	cacheDir, err := getResponseCacheDir()
	if err != nil {
		return nil, err
	}

	return &ResponseCache{
		entries:  make(map[string]*CacheEntry),
		maxSize:  maxSize,
		maxAge:   maxAge,
		cacheDir: cacheDir,
	}, nil
}

// Get retrieves a cached response
func (rc *ResponseCache) Get(key string) (interface{}, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	entry, ok := rc.entries[key]
	if !ok {
		// Try to load from disk
		return rc.loadFromDisk(key)
	}

	if !entry.IsValid() {
		return nil, false
	}

	return entry.Data, true
}

// Set stores a response in cache
func (rc *ResponseCache) Set(key string, data interface{}, ttl time.Duration) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Evict old entries if at capacity
	if len(rc.entries) >= rc.maxSize {
		rc.evictOldest()
	}

	entry := &CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		Key:       key,
		TTL:       ttl,
	}

	rc.entries[key] = entry

	// Persist to disk asynchronously
	go rc.saveToDisk(key, entry)
}

// Invalidate removes an entry from cache
func (rc *ResponseCache) Invalidate(key string) {
	rc.mu.Lock()
	delete(rc.entries, key)
	rc.mu.Unlock()

	// Remove from disk
	rc.deleteFromDisk(key)
}

// InvalidateAll clears all cache
func (rc *ResponseCache) InvalidateAll() {
	rc.mu.Lock()
	rc.entries = make(map[string]*CacheEntry)
	rc.mu.Unlock()

	// Clear disk cache
	os.RemoveAll(rc.cacheDir)
}

// evictOldest removes the oldest entries
func (rc *ResponseCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range rc.entries {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		delete(rc.entries, oldestKey)
	}
}

// generateKey creates a cache key from operation and parameters
func generateKey(operation string, params map[string]interface{}) string {
	data, _ := json.Marshal(params)
	h := sha256.New()
	h.Write([]byte(operation))
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// loadFromDisk loads a cache entry from disk
func (rc *ResponseCache) loadFromDisk(key string) (interface{}, bool) {
	path := filepath.Join(rc.cacheDir, key+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	if !entry.IsValid() {
		os.Remove(path)
		return nil, false
	}

	// Add to memory cache
	rc.mu.Lock()
	rc.entries[key] = &entry
	rc.mu.Unlock()

	return entry.Data, true
}

// saveToDisk persists a cache entry to disk
func (rc *ResponseCache) saveToDisk(key string, entry *CacheEntry) {
	os.MkdirAll(rc.cacheDir, 0755)
	
	path := filepath.Join(rc.cacheDir, key+".json")
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	os.WriteFile(path, data, 0600)
}

// deleteFromDisk removes a cache entry from disk
func (rc *ResponseCache) deleteFromDisk(key string) {
	path := filepath.Join(rc.cacheDir, key+".json")
	os.Remove(path)
}

// getResponseCacheDir returns the cache directory
func getResponseCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "xsh", "responses"), nil
}

// Global response cache instance
var (
	globalResponseCache     *ResponseCache
	globalResponseCacheOnce sync.Once
)

// GetResponseCache returns the global response cache
func GetResponseCache() *ResponseCache {
	globalResponseCacheOnce.Do(func() {
		var err error
		globalResponseCache, err = NewResponseCache(100, 5*time.Minute)
		if err != nil {
			globalResponseCache = &ResponseCache{
				entries: make(map[string]*CacheEntry),
				maxSize: 100,
				maxAge:  5 * time.Minute,
			}
		}
	})
	return globalResponseCache
}

// CachedRequest makes a cached request
func CachedRequest(operation string, params map[string]interface{}, ttl time.Duration, fetchFunc func() (interface{}, error)) (interface{}, error) {
	cache := GetResponseCache()
	key := generateKey(operation, params)

	// Try cache first
	if data, ok := cache.Get(key); ok {
		return data, nil
	}

	// Fetch fresh data
	data, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	// Store in cache
	cache.Set(key, data, ttl)
	return data, nil
}
