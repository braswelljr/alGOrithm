package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Cycle sorts s in ascending order with cycle sort: put each value straight
// into the position it belongs in, following the cycle of displaced values, so
// every value is written at most once (the sort for memory where writes are
// expensive). It sorts in place and returns s. It is not stable.
//
// Time: O(n^2) comparisons in every case; at most n writes.
// Space: O(1).
func Cycle[T cmp.Ordered](s []T) []T { return cycle(s, nop) }

// TraceCycle is Cycle, reporting each comparison with the value in hand
// (Compare with one index) and each write (Set).
func TraceCycle[T cmp.Ordered](s []T, r trace.Recorder) []T { return cycle(s, rec{r}) }

func cycle[T cmp.Ordered](s []T, x rec) []T {
	place := func(start int, item T) int {
		pos := start
		for i := start + 1; i < len(s); i++ {
			x.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{i}})
			if s[i] < item {
				pos++
			}
		}
		return pos
	}
	for start := 0; start < len(s)-1; start++ {
		item := s[start]
		pos := place(start, item)
		if pos == start {
			x.mark("sorted", start)
			continue
		}
		for pos != start {
			for item == s[pos] { // skip past equal values already placed
				pos++
			}
			item, s[pos] = s[pos], item
			set(x, pos, s[pos])
			pos = place(start, item)
		}
		for item == s[pos] && pos != start {
			pos++
		}
		s[start] = item
		set(x, start, item)
		x.mark("sorted", start)
	}
	return s
}
