package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// combShrink is comb sort's gap shrink factor, the value Dobosiewicz and Borowy
// found best.
const combShrink = 1.3

// Comb sorts s in ascending order with comb sort: bubble sort over a gap that
// shrinks by 1.3 each pass, so small values far down the list move early. It
// sorts in place and returns s. It is not stable.
//
// Time: O(n^2) in the worst case, about O(n log n) on random input.
// Space: O(1).
func Comb[T cmp.Ordered](s []T) []T { return comb(s, nop) }

// TraceComb is Comb, reporting each comparison and swap to r.
func TraceComb[T cmp.Ordered](s []T, r trace.Recorder) []T { return comb(s, rec{r}) }

func comb[T cmp.Ordered](s []T, x rec) []T {
	gap := len(s)
	sorted := len(s) < 2
	for !sorted {
		gap = int(float64(gap) / combShrink)
		if gap <= 1 {
			gap, sorted = 1, true // unless this pass swaps something
		}
		for i := 0; i+gap < len(s); i++ {
			x.compare(i, i+gap)
			if s[i] > s[i+gap] {
				s[i], s[i+gap] = s[i+gap], s[i]
				x.swap(i, i+gap)
				sorted = false
			}
		}
	}
	return s
}
