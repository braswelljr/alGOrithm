package searching

import (
	"cmp"
	"math"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Jump returns the index of target in s, which must be sorted: jump ahead in
// blocks of sqrt(n), then scan the one block that can hold the target. -1 when
// absent.
//
// Time: O(sqrt n).
// Space: O(1).
func Jump[T cmp.Ordered](s []T, target T) int { return jump(s, target, nop) }

// TraceJump is Jump, reporting each block (Mark "range") and each element
// compared with the target.
func TraceJump[T cmp.Ordered](s []T, target T, r trace.Recorder) int {
	return jump(s, target, rec{r})
}

func jump[T cmp.Ordered](s []T, target T, x rec) int {
	n := len(s)
	if n == 0 {
		return -1
	}
	step := int(math.Sqrt(float64(n)))
	prev, next := 0, step
	for {
		end := min(next, n) - 1
		x.span(prev, end)
		x.probe(end)
		if s[end] >= target {
			break
		}
		prev, next = next, next+step
		if prev >= n {
			return -1
		}
	}
	for i := prev; i < min(next, n); i++ {
		x.probe(i)
		if s[i] == target {
			return x.found(i)
		}
		if s[i] > target {
			return -1
		}
	}
	return -1
}
