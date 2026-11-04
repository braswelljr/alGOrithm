package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Pancake sorts s in ascending order with pancake sort: the only move is to
// flip a prefix over, like a stack of pancakes under a spatula. Flip the
// largest value to the top, then down into place, and repeat. It sorts in
// place and returns s. It is not stable.
//
// Time: O(n^2), with at most 2n - 3 flips.
// Space: O(1).
func Pancake[T cmp.Ordered](s []T) []T { return pancake(s, nop) }

// TracePancake is Pancake, reporting each comparison, each flip (Mark "flip"
// over the prefix) and the swaps that make it.
func TracePancake[T cmp.Ordered](s []T, r trace.Recorder) []T { return pancake(s, rec{r}) }

func pancake[T cmp.Ordered](s []T, x rec) []T {
	flip := func(k int) {
		if k == 0 {
			return
		}
		x.mark("flip", 0, k)
		for i, j := 0, k; i < j; i, j = i+1, j-1 {
			s[i], s[j] = s[j], s[i]
			x.swap(i, j)
		}
		x.mark("flip")
	}
	for size := len(s); size > 1; size-- {
		largest := 0
		for i := 1; i < size; i++ {
			x.compare(i, largest)
			if s[i] > s[largest] {
				largest = i
			}
		}
		if largest != size-1 {
			flip(largest)
			flip(size - 1)
		}
		x.mark("sorted", size-1)
	}
	return s
}
