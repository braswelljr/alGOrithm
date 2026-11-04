package searching

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Binary returns the index of target in s, which must be sorted, by halving
// the range still searched; -1 when target is absent. With duplicates it
// returns one of their indexes. s is not modified.
//
// Time: O(log n).
// Space: O(1).
func Binary[T cmp.Ordered](s []T, target T) int { return binary(s, target, nop) }

// TraceBinary is Binary, reporting to r as the package describes.
func TraceBinary[T cmp.Ordered](s []T, target T, r trace.Recorder) int {
	return binary(s, target, rec{r})
}

func binary[T cmp.Ordered](s []T, target T, x rec) int {
	return binaryRange(s, target, 0, len(s)-1, x)
}

// binaryRange searches s[lo..hi]; exponential search shares it.
func binaryRange[T cmp.Ordered](s []T, target T, lo, hi int, x rec) int {
	for lo <= hi {
		x.span(lo, hi)
		mid := lo + (hi-lo)/2
		x.probe(mid)
		switch {
		case s[mid] == target:
			return x.found(mid)
		case s[mid] < target:
			lo = mid + 1
		default:
			hi = mid - 1
		}
	}
	return -1
}
