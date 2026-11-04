package greedy_test

import (
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/greedy"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// TestActivitySelectionTextbook is CLRS's example: eleven activities, of which
// at most four fit without overlapping, and the untraced call agrees with the
// traced one.
func TestActivitySelectionTextbook(t *testing.T) {
	acts := []greedy.Interval{
		{"a1", 1, 4}, {"a2", 3, 5}, {"a3", 0, 6}, {"a4", 5, 7}, {"a5", 3, 9}, {"a6", 5, 9},
		{"a7", 6, 10}, {"a8", 8, 11}, {"a9", 8, 12}, {"a10", 2, 14}, {"a11", 12, 16},
	}
	got := greedy.ActivitySelection(acts)
	if len(got) != 4 {
		t.Fatalf("ActivitySelection picked %d activities, want 4: %v", len(got), got)
	}
	// No two chosen activities overlap.
	for i := 0; i < len(got); i++ {
		for j := i + 1; j < len(got); j++ {
			a, b := acts[got[i]], acts[got[j]]
			if a.Start < b.End && b.Start < a.End {
				t.Errorf("%s and %s overlap", a.Name, b.Name)
			}
		}
	}
	var log trace.Log
	if traced := greedy.TraceActivitySelection(acts, &log); !reflect.DeepEqual(traced, got) {
		t.Errorf("traced %v, untraced %v", traced, got)
	}
	if len(greedy.ActivitySelection(nil)) != 0 {
		t.Error("no activities, yet some were chosen")
	}
}
