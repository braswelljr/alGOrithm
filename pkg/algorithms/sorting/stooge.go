package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Stooge sorts s in ascending order with stooge sort: put the ends in order,
// then sort the first two thirds, the last two thirds, and the first two
// thirds again. It is correct and famously slow, worse than bubble sort, a
// lesson in how a recurrence (T(n) = 3T(2n/3) + 1) sets the cost. It sorts in
// place and returns s. It is not stable.
//
// Time: O(n^(log 3 / log 1.5)), about O(n^2.71).
// Space: O(log n) recursion.
func Stooge[T cmp.Ordered](s []T) []T { return stooge(s, nop) }

// TraceStooge is Stooge, reporting each comparison and swap and marking each
// range it recurses on.
func TraceStooge[T cmp.Ordered](s []T, r trace.Recorder) []T { return stooge(s, rec{r}) }

func stooge[T cmp.Ordered](s []T, x rec) []T {
	if len(s) > 1 {
		stoogeRange(s, 0, len(s)-1, x)
	}
	return s
}

func stoogeRange[T cmp.Ordered](s []T, lo, hi int, x rec) {
	x.compare(lo, hi)
	if s[lo] > s[hi] {
		s[lo], s[hi] = s[hi], s[lo]
		x.swap(lo, hi)
	}
	if hi-lo+1 > 2 {
		third := (hi - lo + 1) / 3
		x.mark("range", lo, hi)
		stoogeRange(s, lo, hi-third, x)
		stoogeRange(s, lo+third, hi, x)
		stoogeRange(s, lo, hi-third, x)
	}
}
