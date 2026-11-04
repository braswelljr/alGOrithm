package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Bubble sorts s in ascending order with bubble sort: pass after pass, swap
// neighbors that are out of order, until a pass swaps nothing. It sorts in
// place and returns s. It is stable.
//
// Time: O(n^2) comparisons in the worst and average case, O(n) when s is
// already sorted.
// Space: O(1).
func Bubble[T cmp.Ordered](s []T) []T { return bubble(s, nop) }

// TraceBubble is Bubble, reporting each comparison and swap to r and marking
// each position "sorted" as it settles.
func TraceBubble[T cmp.Ordered](s []T, r trace.Recorder) []T { return bubble(s, rec{r}) }

func bubble[T cmp.Ordered](s []T, x rec) []T {
	for i := 0; i < len(s); i++ {
		swapped := false
		end := len(s) - i - 1
		for j := 0; j < end; j++ {
			x.compare(j, j+1)
			if s[j] > s[j+1] {
				s[j], s[j+1] = s[j+1], s[j]
				x.swap(j, j+1)
				swapped = true
			}
		}
		x.mark("sorted", end)
		if !swapped {
			for k := end - 1; k >= 0; k-- { // everything before end is in order too
				x.mark("sorted", k)
			}
			break
		}
	}
	return s
}
