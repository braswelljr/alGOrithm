package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// timMinRun is the shortest run Tim sort builds with insertion sort before
// merging; Python and Java use 32 to 64.
const timMinRun = 32

// Tim returns a sorted copy of s, in ascending order, made with Tim sort (Tim
// Peters, 2002): find the runs already in order (reversing descending ones),
// extend short runs to a minimum length with insertion sort, then merge runs
// pairwise. Real data is often partly sorted, and Tim sort is then close to
// O(n). This version keeps the run stack simple (merge neighbors bottom up)
// and leaves out galloping. s is not modified. It is stable.
//
// Time: O(n log n) in the worst case, O(n) on sorted or reversed input.
// Space: O(n) for the copy and the merge buffer.
func Tim[T cmp.Ordered](s []T) []T { return tim(s, nop) }

// TraceTim is Tim, reporting each comparison, swap and write, and marking
// each run found ("range") before the merges. Indices refer to the copy being
// sorted.
func TraceTim[T cmp.Ordered](s []T, r trace.Recorder) []T { return tim(s, rec{r}) }

func tim[T cmp.Ordered](in []T, x rec) []T {
	s := clone(in)
	n := len(s)
	var runs [][2]int // [start, end)
	for i := 0; i < n; {
		j := i + 1
		if j < n {
			x.compare(i, j)
			if s[j] < s[i] { // a strictly descending run: reverse it, which keeps stability
				for j+1 < n {
					x.compare(j, j+1)
					if s[j+1] >= s[j] {
						break
					}
					j++
				}
				for a, b := i, j; a < b; a, b = a+1, b-1 {
					s[a], s[b] = s[b], s[a]
					x.swap(a, b)
				}
				j++
			} else {
				for j+1 < n {
					x.compare(j, j+1)
					if s[j+1] < s[j] {
						break
					}
					j++
				}
				j++
			}
		}
		end := min(max(j, i+timMinRun), n)
		insertionRange(s, i, end, x)
		x.mark("range", i, end-1)
		runs = append(runs, [2]int{i, end})
		i = end
	}
	buf := make([]T, n)
	for len(runs) > 1 {
		var next [][2]int
		for k := 0; k+1 < len(runs); k += 2 {
			lo, mid, hi := runs[k][0], runs[k][1], runs[k+1][1]
			mergeRuns(s, buf, lo, mid, hi, x)
			next = append(next, [2]int{lo, hi})
		}
		if len(runs)%2 == 1 {
			next = append(next, runs[len(runs)-1])
		}
		runs = next
	}
	return s
}

// insertionRange insertion-sorts s[lo:hi], knowing a prefix may already be in
// order.
func insertionRange[T cmp.Ordered](s []T, lo, hi int, x rec) {
	for i := lo + 1; i < hi; i++ {
		for j := i; j > lo; j-- {
			x.compare(j-1, j)
			if s[j-1] <= s[j] {
				break
			}
			s[j-1], s[j] = s[j], s[j-1]
			x.swap(j-1, j)
		}
	}
}

// mergeRuns merges the sorted runs s[lo:mid] and s[mid:hi], stably.
func mergeRuns[T cmp.Ordered](s, buf []T, lo, mid, hi int, x rec) {
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
			if buf[i] <= buf[j] {
				v, i = buf[i], i+1
			} else {
				v, j = buf[j], j+1
			}
		}
		s[k] = v
		set(x, k, v)
	}
}
