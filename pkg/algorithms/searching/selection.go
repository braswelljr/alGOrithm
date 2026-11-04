package searching

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Quickselect returns the kth smallest value of s (k from 0, so k = 0 is the
// minimum and k = len(s)/2 a median) with Tony Hoare's quickselect: partition
// around a pivot like quick sort, then continue only into the side that holds
// position k. It reports false when k is out of range. It works on a copy, so
// s is not modified. The pivot is the middle element; a random pivot (see
// package randomized) protects against crafted input.
//
// Time: O(n) on average, O(n^2) in the worst case.
// Space: O(n) for the copy.
func Quickselect[T cmp.Ordered](s []T, k int) (T, bool) {
	return quickselect(s, k, nop, middlePivot[T])
}

// TraceQuickselect is Quickselect, reporting each range still searched (Mark
// "range"), each comparison with the pivot (Compare, two indexes), each swap
// (Swap) and the final position (Mark "found"). Indices refer to the copy, so
// trace.Replay of the swaps on s rebuilds its partitioned state.
func TraceQuickselect[T cmp.Ordered](s []T, k int, r trace.Recorder) (T, bool) {
	return quickselect(s, k, rec{r}, middlePivot[T])
}

// MedianOfMedians returns the kth smallest value of s (k from 0) in
// guaranteed linear time with the median-of-medians pivot (Blum, Floyd,
// Pratt, Rivest and Tarjan, 1973): split the range into groups of five, take
// each group's median, and pick the median of those (recursively) as the
// pivot. That pivot always has at least 30% of the values on each side, so
// the range shrinks geometrically. It reports false when k is out of range. s
// is not modified.
//
// Time: O(n) in the worst case, with a larger constant than Quickselect.
// Space: O(n).
func MedianOfMedians[T cmp.Ordered](s []T, k int) (T, bool) {
	return quickselect(s, k, nop, medianOfMediansPivot[T])
}

// TraceMedianOfMedians is MedianOfMedians, reporting to r like
// TraceQuickselect.
func TraceMedianOfMedians[T cmp.Ordered](s []T, k int, r trace.Recorder) (T, bool) {
	return quickselect(s, k, rec{r}, medianOfMediansPivot[T])
}

// pivotFunc returns the index of the pivot to use for s[lo..hi].
type pivotFunc[T cmp.Ordered] func(s []T, lo, hi int) int

func middlePivot[T cmp.Ordered](_ []T, lo, hi int) int { return lo + (hi-lo)/2 }

func medianOfMediansPivot[T cmp.Ordered](s []T, lo, hi int) int {
	var medians []T
	for g := lo; g <= hi; g += 5 {
		group := append([]T(nil), s[g:min(g+5, hi+1)]...)
		for i := 1; i < len(group); i++ { // insertion sort of at most five
			for j := i; j > 0 && group[j] < group[j-1]; j-- {
				group[j], group[j-1] = group[j-1], group[j]
			}
		}
		medians = append(medians, group[(len(group)-1)/2])
	}
	m, _ := quickselect(medians, (len(medians)-1)/2, nop, medianOfMediansPivot[T])
	for i := lo; i <= hi; i++ {
		if s[i] == m {
			return i
		}
	}
	return lo // unreachable: m is one of s[lo..hi]
}

func quickselect[T cmp.Ordered](in []T, k int, x rec, pivot pivotFunc[T]) (T, bool) {
	var zero T
	if k < 0 || k >= len(in) {
		return zero, false
	}
	s := append([]T(nil), in...)
	lo, hi := 0, len(s)-1
	for lo < hi {
		x.span(lo, hi)
		p := pivot(s, lo, hi)
		p = partitionAround(s, lo, hi, p, x)
		switch {
		case k == p:
			x.found(p)
			return s[p], true
		case k < p:
			hi = p - 1
		default:
			lo = p + 1
		}
	}
	x.found(lo)
	return s[lo], true
}

// partitionAround moves s[p] to its sorted place in s[lo..hi] (Lomuto
// partition), smaller values before it, and returns that place.
func partitionAround[T cmp.Ordered](s []T, lo, hi, p int, x rec) int {
	swap := func(i, j int) {
		if i != j {
			s[i], s[j] = s[j], s[i]
			x.r.Record(trace.Event{Kind: trace.Swap, Indices: []int{i, j}})
		}
	}
	swap(p, hi)
	store := lo
	for i := lo; i < hi; i++ {
		x.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{i, hi}})
		if s[i] < s[hi] {
			swap(i, store)
			store++
		}
	}
	swap(store, hi)
	return store
}
