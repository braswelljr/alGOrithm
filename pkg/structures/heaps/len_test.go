package heaps_test

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
)

// TestLen checks that both heaps count what they hold through pushes and pops.
func TestLen(t *testing.T) {
	h := heaps.NewMinHeap(5, 3, 8)
	if h.Len() != 3 {
		t.Fatalf("NewMinHeap(5, 3, 8).Len() = %d", h.Len())
	}
	h.Push(1)
	if h.Len() != 4 {
		t.Errorf("after a push Len() = %d, want 4", h.Len())
	}
	if v, ok := h.Pop(); !ok || v != 1 || h.Len() != 3 {
		t.Errorf("Pop() = %d, %v with Len() %d; want 1, true, 3", v, ok, h.Len())
	}
	p := heaps.NewPriorityHeap(func(a, b string) bool { return len(a) < len(b) }, "pear", "fig")
	p.Push("banana")
	if p.Len() != 3 {
		t.Errorf("PriorityHeap Len() = %d, want 3", p.Len())
	}
	if v, ok := p.Pop(); !ok || v != "fig" || p.Len() != 2 {
		t.Errorf("PriorityHeap Pop() = %q, %v with Len() %d; want fig, true, 2", v, ok, p.Len())
	}
}
