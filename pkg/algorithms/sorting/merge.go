package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Merge returns a sorted copy of s, in ascending order, made with merge sort
// (John von Neumann, 1945): sort each half, then merge the two sorted halves.
// s is not modified. It is stable.
//
// Time: O(n log n) in every case.
// Space: O(n) for the copy and the merge buffer.
func Merge[T cmp.Ordered](s []T) []T { return mergeSort(s, nop) }

// TraceMerge is Merge, reporting each comparison and each value written back
// during a merge (Set), and marking the range being merged. Indices refer to
// the copy being sorted, so replaying the events on a copy of s gives the
// result.
func TraceMerge[T cmp.Ordered](s []T, r trace.Recorder) []T { return mergeSort(s, rec{r}) }

func mergeSort[T cmp.Ordered](in []T, x rec) []T {
	s := clone(in)
	mergeRange(s, make([]T, len(s)), 0, len(s), x)
	return s
}

// mergeRange sorts s[lo:hi], using buf[lo:hi] as scratch space.
func mergeRange[T cmp.Ordered](s, buf []T, lo, hi int, x rec) {
	if hi-lo < 2 {
		return
	}
	mid := lo + (hi-lo)/2
	mergeRange(s, buf, lo, mid, x)
	mergeRange(s, buf, mid, hi, x)
	x.mark("range", lo, hi-1)
	copy(buf[lo:hi], s[lo:hi])
	i, j := lo, mid
	for k := lo; k < hi; k++ {
		var v T
		switch {
		case i == mid:
			v, j = buf[j], j+1
		case j == hi:
			v, i = buf[i], i+1
		default:
			x.compare(i, j)
			if buf[i] <= buf[j] { // a tie takes the left value: stable
				v, i = buf[i], i+1
			} else {
				v, j = buf[j], j+1
			}
		}
		s[k] = v
		set(x, k, v)
	}
}
