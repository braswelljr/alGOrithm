package lookup_test

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
)

// TestCapacityAndLen checks the caches' sizes: Capacity is what they were made
// with, and the TTL cache's Len counts the keys it holds, never more than its
// capacity.
func TestCapacityAndLen(t *testing.T) {
	if c := lookup.NewLRU[string, int](3); c.Capacity() != 3 {
		t.Errorf("LRU Capacity() = %d, want 3", c.Capacity())
	}
	if c := lookup.NewLFUCache[string, int](2); c.Capacity() != 2 {
		t.Errorf("LFU Capacity() = %d, want 2", c.Capacity())
	}
	c := lookup.NewTTLCache[string, int](2, 10)
	if c.Len() != 0 {
		t.Fatalf("a new TTL cache holds %d", c.Len())
	}
	c.Put("a", 1, 0)
	c.Put("b", 2, 1)
	c.Put("c", 3, 2)
	if c.Len() != 2 {
		t.Errorf("Len() = %d after three puts into two slots, want 2", c.Len())
	}
}
