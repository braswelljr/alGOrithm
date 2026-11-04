package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Insertion sorts s in ascending order with insertion sort: take each value in
// turn and move it left past every larger value before it. It sorts in place
// and returns s. It is stable.
//
// Time: O(n^2) in the worst and average case, O(n) when s is already sorted.
// Space: O(1).
func Insertion[T cmp.Ordered](s []T) []T { return insertion(s, nop) }

// TraceInsertion is Insertion, reporting each comparison and swap to r.
func TraceInsertion[T cmp.Ordered](s []T, r trace.Recorder) []T { return insertion(s, rec{r}) }

func insertion[T cmp.Ordered](s []T, x rec) []T {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0; j-- {
			x.compare(j-1, j)
			if s[j-1] <= s[j] {
				break // s[:i+1] is in order; equal values keep their order
			}
			s[j-1], s[j] = s[j], s[j-1]
			x.swap(j-1, j)
		}
	}
	return s
}
