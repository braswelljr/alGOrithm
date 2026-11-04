package searching

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Interpolation returns the index of target in s, which must be sorted, by
// guessing its position from its value, as a person opens a phone book near
// the right letter. -1 when absent.
//
// Time: O(log log n) on uniformly distributed values, O(n) in the worst case.
// Space: O(1).
func Interpolation(s []int, target int) int { return interpolation(s, target, nop) }

// TraceInterpolation is Interpolation, reporting to r as the package
// describes.
func TraceInterpolation(s []int, target int, r trace.Recorder) int {
	return interpolation(s, target, rec{r})
}

func interpolation(s []int, target int, x rec) int {
	lo, hi := 0, len(s)-1
	// Outside [s[lo], s[hi]] the target cannot be present, and the estimate
	// below would fall outside the range.
	for lo <= hi && target >= s[lo] && target <= s[hi] {
		x.span(lo, hi)
		pos := lo
		if s[hi] != s[lo] { // all equal: probe lo, never divide by zero
			pos = lo + int(int64(hi-lo)*(int64(target)-int64(s[lo]))/(int64(s[hi])-int64(s[lo])))
		}
		x.probe(pos)
		switch {
		case s[pos] == target:
			return x.found(pos)
		case s[pos] < target:
			lo = pos + 1
		default:
			hi = pos - 1
		}
	}
	return -1
}
