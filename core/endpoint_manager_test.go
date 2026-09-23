package core

import "testing"

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
