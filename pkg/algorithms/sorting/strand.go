package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Strand returns a sorted copy of s, in ascending order, made with strand
// sort: pull an increasing "strand" out of the remaining values (take the
// first, then every later value at least as large as the last taken), merge
// it into the result, and repeat until nothing remains. It is fast when the
// input has few long increasing runs. s is not modified. It is stable.
//
// Time: O(n^2) in the worst case, O(n) when s is already sorted.
// Space: O(n).
func Strand[T cmp.Ordered](s []T) []T { return strand(s, nop) }

// TraceStrand is Strand, reporting each value compared while pulling a strand
// (Compare with the original indexes), each value taken (Visit), and finally
// each value of the result (Set, the result written over a copy of s).
func TraceStrand[T cmp.Ordered](s []T, r trace.Recorder) []T { return strand(s, rec{r}) }

func strand[T cmp.Ordered](in []T, x rec) []T {
	type item struct {
		v T
		i int // original index, for the trace
	}
	rest := make([]item, len(in))
	for i, v := range in {
		rest[i] = item{v, i}
	}
	var result []T
	for len(rest) > 0 {
		strandItems := []item{rest[0]}
		x.visit(rest[0].i)
		var left []item
		for _, it := range rest[1:] {
			last := strandItems[len(strandItems)-1]
			x.compare(last.i, it.i)
			if it.v >= last.v {
				strandItems = append(strandItems, it)
				x.visit(it.i)
			} else {
				left = append(left, it)
			}
		}
		merged := make([]T, 0, len(result)+len(strandItems))
		a := 0
		for _, it := range strandItems {
			for a < len(result) && result[a] <= it.v { // result first on ties: stable
				merged = append(merged, result[a])
				a++
			}
			merged = append(merged, it.v)
		}
		result = append(merged, result[a:]...)
		rest = left
	}
	out := clone(in)
	for i, v := range result {
		out[i] = v
		set(x, i, v)
	}
	return out
}
