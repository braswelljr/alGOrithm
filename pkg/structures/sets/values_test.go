package sets_test

import (
	"slices"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/sets"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// TestOrderedSetValues checks that Values are sorted, distinct, and a copy the
// caller may change without changing the set.
func TestOrderedSetValues(t *testing.T) {
	s := sets.NewOrderedSet(5, 1, 3, 1)
	got := s.Values()
	if !slices.Equal(got, []int{1, 3, 5}) {
		t.Fatalf("Values() = %v, want [1 3 5]", got)
	}
	got[0] = 99
	if s.Values()[0] != 1 {
		t.Error("changing the returned slice changed the set")
	}
}

// TestTracePowerSet checks the traced power set against the untraced one: 2^n
// subsets, the same ones, and events reported.
func TestTracePowerSet(t *testing.T) {
	want, err := sets.PowerSet([]int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	var log trace.Log
	got, err := sets.TracePowerSet([]int{1, 2, 3}, &log)
	if err != nil || len(got) != 8 || len(got) != len(want) || log.Total() == 0 {
		t.Errorf("TracePowerSet gave %d subsets and %d events (%v), want 8 and some", len(got), log.Total(), err)
	}
}
