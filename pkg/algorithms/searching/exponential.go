package searching

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Exponential returns the index of target in s, which must be sorted: double a
// bound (1, 2, 4, ...) until it passes the target, then binary search the last
// range. It suits lists whose length is unknown or unbounded. -1 when absent.
//
// Time: O(log i) for a target at index i.
// Space: O(1).
func Exponential[T cmp.Ordered](s []T, target T) int { return exponential(s, target, nop) }

// TraceExponential is Exponential, reporting each bound probed and then what
// TraceBinary reports.
func TraceExponential[T cmp.Ordered](s []T, target T, r trace.Recorder) int {
	return exponential(s, target, rec{r})
}

func exponential[T cmp.Ordered](s []T, target T, x rec) int {
	if len(s) == 0 {
		return -1
	}
	bound := 1
	for bound < len(s) {
		x.probe(bound)
		if s[bound] >= target {
			break
		}
		bound *= 2
	}
	return binaryRange(s, target, bound/2, min(bound, len(s)-1), x)
}
