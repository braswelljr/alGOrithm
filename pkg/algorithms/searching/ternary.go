package searching

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Ternary returns the index of target in s, which must be sorted, splitting the
// range into thirds at two midpoints and keeping the third that can hold the
// target. -1 when absent.
//
// Time: O(log n) (log base 3, but two comparisons a step make it slower than
// binary search in practice).
// Space: O(1).
func Ternary[T cmp.Ordered](s []T, target T) int { return ternary(s, target, nop) }

// TraceTernary is Ternary, reporting to r as the package describes.
func TraceTernary[T cmp.Ordered](s []T, target T, r trace.Recorder) int {
	return ternary(s, target, rec{r})
}

func ternary[T cmp.Ordered](s []T, target T, x rec) int {
	lo, hi := 0, len(s)-1
	for lo <= hi {
		x.span(lo, hi)
		third := (hi - lo) / 3
		m1, m2 := lo+third, hi-third
		x.probe(m1)
		if s[m1] == target {
			return x.found(m1)
		}
		x.probe(m2)
		if s[m2] == target {
			return x.found(m2)
		}
		switch {
		case target < s[m1]:
			hi = m1 - 1
		case target > s[m2]:
			lo = m2 + 1
		default:
			lo, hi = m1+1, m2-1
		}
	}
	return -1
}
