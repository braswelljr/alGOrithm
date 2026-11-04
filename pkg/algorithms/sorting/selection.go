package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Selection sorts s in ascending order with selection sort: find the smallest
// remaining value and swap it into the next position. It sorts in place and
// returns s. It is not stable.
//
// Time: O(n^2) comparisons in every case, but at most n-1 swaps.
// Space: O(1).
func Selection[T cmp.Ordered](s []T) []T { return selection(s, nop) }

// TraceSelection is Selection, reporting each comparison and swap to r,
// marking the smallest value so far "min" and each settled position "sorted".
func TraceSelection[T cmp.Ordered](s []T, r trace.Recorder) []T { return selection(s, rec{r}) }

func selection[T cmp.Ordered](s []T, x rec) []T {
	for i := 0; i < len(s); i++ {
		least := i
		x.mark("min", least)
		for j := i + 1; j < len(s); j++ {
			x.compare(j, least)
			if s[j] < s[least] {
				least = j
				x.mark("min", least)
			}
		}
		if least != i {
			s[i], s[least] = s[least], s[i]
			x.swap(i, least)
		}
		x.mark("sorted", i)
	}
	x.mark("min")
	return s
}
