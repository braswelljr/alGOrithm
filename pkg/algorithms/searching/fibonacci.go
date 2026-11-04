package searching

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Fibonacci returns the index of target in s, which must be sorted, splitting
// the range at Fibonacci numbers instead of halves, so the split points need
// only additions and subtractions. -1 when absent.
//
// Time: O(log n).
// Space: O(1).
func Fibonacci[T cmp.Ordered](s []T, target T) int { return fibonacci(s, target, nop) }

// TraceFibonacci is Fibonacci, reporting to r as the package describes.
func TraceFibonacci[T cmp.Ordered](s []T, target T, r trace.Recorder) int {
	return fibonacci(s, target, rec{r})
}

func fibonacci[T cmp.Ordered](s []T, target T, x rec) int {
	n := len(s)
	a, b := 0, 1 // F(k-2), F(k-1)
	c := a + b   // F(k): the smallest Fibonacci number >= n
	for c < n {
		a, b, c = b, c, b+c
	}
	offset := -1
	for c > 1 {
		i := min(offset+a, n-1)
		x.span(offset+1, n-1)
		x.probe(i)
		switch {
		case s[i] < target: // keep the part after i
			c, b, a = b, a, b-a
			offset = i
		case s[i] > target: // keep the part before i
			c, b, a = a, b-a, a-(b-a)
		default:
			return x.found(i)
		}
	}
	if b == 1 && offset+1 < n {
		x.probe(offset + 1)
		if s[offset+1] == target {
			return x.found(offset + 1)
		}
	}
	return -1
}
