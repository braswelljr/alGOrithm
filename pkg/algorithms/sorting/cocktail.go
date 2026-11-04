package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Cocktail sorts s in ascending order with cocktail shaker sort: bubble sort
// passing forward and then backward, so small values near the end move left as
// fast as large values move right. It sorts in place and returns s. It is
// stable.
//
// Time: O(n^2), O(n) when s is already sorted.
// Space: O(1).
func Cocktail[T cmp.Ordered](s []T) []T { return cocktail(s, nop) }

// TraceCocktail is Cocktail, reporting each comparison and swap to r and
// marking each end "sorted" as it settles.
func TraceCocktail[T cmp.Ordered](s []T, r trace.Recorder) []T { return cocktail(s, rec{r}) }

func cocktail[T cmp.Ordered](s []T, x rec) []T {
	lo, hi := 0, len(s)-1
	step := func(i int) bool {
		x.compare(i, i+1)
		if s[i] > s[i+1] {
			s[i], s[i+1] = s[i+1], s[i]
			x.swap(i, i+1)
			return true
		}
		return false
	}
	for lo < hi {
		swapped := false
		for i := lo; i < hi; i++ {
			swapped = step(i) || swapped
		}
		x.mark("sorted", hi)
		hi--
		for i := hi - 1; i >= lo; i-- {
			swapped = step(i) || swapped
		}
		x.mark("sorted", lo)
		lo++
		if !swapped {
			break
		}
	}
	return s
}
