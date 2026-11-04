package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// OddEven sorts s in ascending order with odd-even transposition sort:
// alternate passes compare the pairs at even and at odd positions. Every
// comparison in a pass is independent, which is why it suits parallel
// hardware. It sorts in place and returns s. It is stable.
//
// Time: O(n^2).
// Space: O(1).
func OddEven[T cmp.Ordered](s []T) []T { return oddEven(s, nop) }

// TraceOddEven is OddEven, reporting each comparison and swap to r.
func TraceOddEven[T cmp.Ordered](s []T, r trace.Recorder) []T { return oddEven(s, rec{r}) }

func oddEven[T cmp.Ordered](s []T, x rec) []T {
	for sorted := len(s) < 2; !sorted; {
		sorted = true
		for start := 0; start <= 1; start++ {
			for i := start; i+1 < len(s); i += 2 {
				x.compare(i, i+1)
				if s[i] > s[i+1] {
					s[i], s[i+1] = s[i+1], s[i]
					x.swap(i, i+1)
					sorted = false
				}
			}
		}
	}
	return s
}
