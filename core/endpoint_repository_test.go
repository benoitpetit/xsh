package core

import (
	"sync"
	"testing"
)

func TestEndpointRepositorySnapshotIsolated(t *testing.T) {
	repository := NewEndpointRepository(nil)
	if err := repository.Replace(&EndpointCache{
		Endpoints:   map[string]string{"SearchTimeline": "query/SearchTimeline"},
		Quarantined: map[string]string{},
		Features:    map[string]bool{"feature": true},
		OpFeatures:  map[string][]string{"SearchTimeline": {"feature"}},
	}); err != nil {
		t.Fatal(err)
	}

	snapshot := repository.Snapshot()
	snapshot.Endpoints["SearchTimeline"] = "mutated"
	snapshot.Features["feature"] = false
	snapshot.OpFeatures["SearchTimeline"][0] = "mutated"

	fresh := repository.Snapshot()
	if fresh.Endpoints["SearchTimeline"] != "query/SearchTimeline" || !fresh.Features["feature"] || fresh.OpFeatures["SearchTimeline"][0] != "feature" {
		t.Fatalf("repository snapshot was mutable: %#v", fresh)
	}
}

func TestEndpointRepositorySerializesMutations(t *testing.T) {
	repository := NewEndpointRepository(nil)
	if err := repository.Replace(&EndpointCache{Features: map[string]bool{"count": false}}); err != nil {
		t.Fatal(err)
	}

	const workers = 32
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := repository.Mutate(func(cache *EndpointCache) error {
				cache.Features["count"] = !cache.Features["count"]
				return nil
			}); err != nil {
				t.Errorf("Mutate() error = %v", err)
			}
		}()
	}
	wg.Wait()
	if repository.Snapshot().Features["count"] {
		t.Fatal("serialized even mutations should preserve the initial false value")
	}
}

func TestEndpointRepositoryPreservesQuarantineUntilReplacement(t *testing.T) {
	repository := NewEndpointRepository(nil)
	if err := repository.Replace(&EndpointCache{
		Endpoints:   map[string]string{"Stale": "old/Stale"},
		Quarantined: map[string]string{"Stale": "HTTP 404"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := repository.Replace(&EndpointCache{Endpoints: map[string]string{"Other": "new/Other"}}); err != nil {
		t.Fatal(err)
	}
	if _, ok := repository.Snapshot().Quarantined["Stale"]; !ok {
		t.Fatal("quarantine should survive a refresh without a replacement")
	}
	if err := repository.Replace(&EndpointCache{Endpoints: map[string]string{"Stale": "new/Stale"}}); err != nil {
		t.Fatal(err)
	}
	if _, ok := repository.Snapshot().Quarantined["Stale"]; ok {
		t.Fatal("quarantine should clear when discovery supplies a replacement")
	}
}
