package sorting_test

import (
	"slices"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
)

// TestStoogeUntraced checks the untraced Stooge on small inputs (it is too
// slow for the shared suite's large ones): sorted, in place, and the same
// slice returned.
func TestStoogeUntraced(t *testing.T) {
	for _, in := range [][]int{nil, {1}, {2, 1}, {3, 1, 2}, {5, 5, 1, 9, -3, 0, 7}, {9, 8, 7, 6, 5, 4, 3, 2, 1, 0}} {
		s := slices.Clone(in)
		got := sorting.Stooge(s)
		want := slices.Clone(in)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Errorf("Stooge(%v) = %v, want %v", in, got, want)
		}
		if len(s) > 0 && &got[0] != &s[0] {
			t.Errorf("Stooge(%v) did not sort in place", in)
		}
	}
	if got := sorting.Stooge([]string{"pear", "fig", "apple"}); !slices.Equal(got, []string{"apple", "fig", "pear"}) {
		t.Errorf("Stooge on strings = %v", got)
	}
}
