package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// shellGaps is Ciura's sequence, the best known for n up to a few thousand.
var shellGaps = []int{701, 301, 132, 57, 23, 10, 4, 1}

// Shell sorts s in ascending order with Shell sort (Donald Shell, 1959):
// insertion sort over a shrinking sequence of gaps, so values far from their
// place move early. It sorts in place and returns s. It is not stable.
//
// Time: about O(n^(4/3)) with these gaps; O(n^2) in the worst case.
// Space: O(1).
func Shell[T cmp.Ordered](s []T) []T { return shell(s, nop) }

// TraceShell is Shell, reporting each comparison and swap to r.
func TraceShell[T cmp.Ordered](s []T, r trace.Recorder) []T { return shell(s, rec{r}) }

func shell[T cmp.Ordered](s []T, x rec) []T {
	for _, gap := range shellGaps {
		for i := gap; i < len(s); i++ {
			for j := i; j >= gap; j -= gap {
				x.compare(j-gap, j)
				if s[j-gap] <= s[j] {
					break
				}
				s[j-gap], s[j] = s[j], s[j-gap]
				x.swap(j-gap, j)
			}
		}
	}
	return s
}
