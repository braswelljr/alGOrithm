package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Quick returns a sorted copy of s, in ascending order, made with quick sort
// (Tony Hoare, 1959): move everything smaller than a pivot to its left and
// everything larger to its right, then sort each side. The pivot is the middle
// value, so sorted and reversed input do not hit the worst case. s is not
// modified. It is not stable.
//
// Time: O(n log n) on average, O(n^2) in the worst case.
// Space: O(n) for the copy, plus O(log n) recursion.
func Quick[T cmp.Ordered](s []T) []T { return quick(s, nop) }

// TraceQuick is Quick, reporting each comparison and swap to r and marking the
// pivot and each settled position. Indices refer to the copy being sorted, so
// replaying the events on a copy of s gives the result.
func TraceQuick[T cmp.Ordered](s []T, r trace.Recorder) []T { return quick(s, rec{r}) }

func quick[T cmp.Ordered](in []T, x rec) []T {
	s := clone(in)
	quickRange(s, 0, len(s)-1, x)
	return s
}

// quickRange sorts s[lo..hi] (inclusive) with Lomuto partitioning.
func quickRange[T cmp.Ordered](s []T, lo, hi int, x rec) {
	for lo < hi {
		p := partition(s, lo, hi, x)
		// Recurse into the smaller side and loop on the larger: the recursion
		// stays O(log n) deep.
		if p-lo < hi-p {
			quickRange(s, lo, p-1, x)
			lo = p + 1
		} else {
			quickRange(s, p+1, hi, x)
			hi = p - 1
		}
	}
	if lo == hi {
		x.mark("sorted", lo)
	}
}

func partition[T cmp.Ordered](s []T, lo, hi int, x rec) int {
	swap := func(i, j int) {
		if i != j {
			s[i], s[j] = s[j], s[i]
			x.swap(i, j)
		}
	}
	swap(lo+(hi-lo)/2, hi) // the pivot waits at the end
	x.mark("pivot", hi)
	store := lo
	for i := lo; i < hi; i++ {
		x.compare(i, hi)
		if s[i] < s[hi] {
			swap(i, store)
			store++
		}
	}
	swap(store, hi)
	x.mark("pivot")
	x.mark("sorted", store)
	return store
}
