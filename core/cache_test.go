package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResponseCacheGetLoadsValidDiskEntryWithoutDeadlocking(t *testing.T) {
	cacheDir := t.TempDir()
	cache := &ResponseCache{entries: make(map[string]*CacheEntry), cacheDir: cacheDir}
	entry := CacheEntry{
		Data:      map[string]interface{}{"value": "cached"},
		Timestamp: time.Now(),
		Key:       "disk-entry",
		TTL:       time.Minute,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, entry.Key+".json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	type result struct {
		data interface{}
		ok   bool
	}
	completed := make(chan result, 1)
	go func() {
		got, ok := cache.Get(entry.Key)
		completed <- result{data: got, ok: ok}
	}()

	select {
	case got := <-completed:
		if !got.ok {
			t.Fatal("Get() did not return the valid disk entry")
		}
		value := got.data.(map[string]interface{})["value"]
		if value != "cached" {
			t.Fatalf("cached value = %#v, want cached", value)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Get() deadlocked while loading a valid disk entry")
	}
}
