package searching

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// LowerBound returns the first index i of s, which must be sorted, with
// s[i] >= target: where target would be inserted to keep s sorted, before
// any equal values. It returns len(s) when every value is smaller. Unlike
// Binary it never fails, and with duplicates it finds the first one (C++'s
// std::lower_bound, Go's sort.Search). s is not modified.
//
// Time: O(log n).
// Space: O(1).
func LowerBound[T cmp.Ordered](s []T, target T) int {
	return bound(s, func(v T) bool { return v >= target }, nop)
}

// TraceLowerBound is LowerBound, reporting to r as the package describes; the
// found mark is the answer, which may be len(s).
func TraceLowerBound[T cmp.Ordered](s []T, target T, r trace.Recorder) int {
	return bound(s, func(v T) bool { return v >= target }, rec{r})
}

// UpperBound returns the first index i of s, which must be sorted, with
// s[i] > target: the insertion point after any equal values. It returns
// len(s) when no value is larger. UpperBound - LowerBound counts the copies of
// target.
//
// Time: O(log n).
// Space: O(1).
func UpperBound[T cmp.Ordered](s []T, target T) int {
	return bound(s, func(v T) bool { return v > target }, nop)
}

// TraceUpperBound is UpperBound, reporting to r like TraceLowerBound.
func TraceUpperBound[T cmp.Ordered](s []T, target T, r trace.Recorder) int {
	return bound(s, func(v T) bool { return v > target }, rec{r})
}

// bound returns the first index where ok holds, ok being false then true
// along s.
func bound[T any](s []T, ok func(T) bool, x rec) int {
	lo, hi := 0, len(s) // the answer is in [lo, hi]
	for lo < hi {
		x.span(lo, hi-1)
		mid := lo + (hi-lo)/2
		x.probe(mid)
		if ok(s[mid]) {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return x.found(lo)
}
