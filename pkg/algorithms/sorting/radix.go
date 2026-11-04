package sorting

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Radix returns a sorted copy of s, in ascending order, made with
// least-significant-digit radix sort: sort by the last decimal digit, then the
// next, each pass a stable counting sort, so no two values are ever compared.
// Negative numbers are handled by shifting every value by the minimum. s is
// not modified. It is stable.
//
// Time: O(n * k), where k is the number of decimal digits of the largest
// shifted value.
// Space: O(n + 10).
func Radix(s []int) []int { return radix(s, nop) }

// TraceRadix is Radix, reporting each value read (Visit), each value written
// back after a pass (Set), and the digit being sorted on (Mark "digit",
// Values[0] the place: 1, 10, 100, ...). Indices refer to a copy of s.
func TraceRadix(s []int, r trace.Recorder) []int { return radix(s, rec{r}) }

func radix(in []int, x rec) []int {
	s := clone(in)
	if len(s) < 2 {
		return s
	}
	lo, hi := s[0], s[0]
	for _, v := range s {
		lo, hi = min(lo, v), max(hi, v)
	}
	digit := func(v int, place uint64) int {
		return int((uint64(int64(v)-int64(lo)) / place) % 10)
	}
	span := uint64(int64(hi) - int64(lo)) // fits in uint64 even when it does not fit in int
	out := make([]int, len(s))
	for place := uint64(1); place <= span; place *= 10 {
		x.r.Record(trace.Event{Kind: trace.Mark, Values: []int{int(place)}, Note: "digit"})
		var count [10]int
		for i, v := range s {
			x.visit(i)
			count[digit(v, place)]++
		}
		for d := 1; d < 10; d++ {
			count[d] += count[d-1]
		}
		for i := len(s) - 1; i >= 0; i-- { // backward keeps each pass stable
			d := digit(s[i], place)
			count[d]--
			out[count[d]] = s[i]
		}
		for i, v := range out {
			s[i] = v
			set(x, i, v)
		}
		if place > span/10 {
			break // the next place would overflow, and there is no digit there
		}
	}
	return s
}
