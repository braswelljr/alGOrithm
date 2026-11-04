package searching

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Linear returns the index of the first element of s equal to target, looking
// at every element in order; -1 when there is none. s need not be sorted.
//
// Time: O(n).
// Space: O(1).
func Linear[T comparable](s []T, target T) int { return linear(s, target, nop) }

// TraceLinear is Linear, reporting each element compared and the index found.
func TraceLinear[T comparable](s []T, target T, r trace.Recorder) int {
	return linear(s, target, rec{r})
}

func linear[T comparable](s []T, target T, x rec) int {
	for i, v := range s {
		x.probe(i)
		if v == target {
			return x.found(i)
		}
	}
	return -1
}
