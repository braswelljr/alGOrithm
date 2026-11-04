package arrays

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Reverse reverses s in place with two pointers walking inward, and returns
// it.
//
// Time: O(n). Space: O(1).
func Reverse(s []int) []int { return reverse(s, 0, len(s)-1, nop) }

// TraceReverse is Reverse, reporting each swap.
func TraceReverse(s []int, r trace.Recorder) []int { return reverse(s, 0, len(s)-1, rec{r}) }

func reverse(s []int, i, j int, x rec) []int {
	for ; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
		x.swap(i, j)
	}
	return s
}

// Rotate rotates s right by k places in place (k may be negative or larger
// than len(s)) with three reversals: reverse all, then each part. It returns
// s.
//
// Time: O(n). Space: O(1).
func Rotate(s []int, k int) []int { return rotate(s, k, nop) }

// TraceRotate is Rotate, reporting each reversal (Mark "range") and swap.
func TraceRotate(s []int, k int, r trace.Recorder) []int { return rotate(s, k, rec{r}) }

func rotate(s []int, k int, x rec) []int {
	n := len(s)
	if n == 0 {
		return s
	}
	k = ((k % n) + n) % n
	for _, part := range [][2]int{{0, n - 1}, {0, k - 1}, {k, n - 1}} {
		if part[0] < part[1] {
			x.mark("range", part[0], part[1])
			reverse(s, part[0], part[1], x)
		}
	}
	return s
}

// TwoSum finds two indexes i < j of the sorted slice s whose values sum to
// target, moving a pointer in from each end: if the sum is too small, move the
// left one right; too big, the right one left. ok is false when no pair sums
// to target.
//
// Time: O(n). Space: O(1).
func TwoSum(s []int, target int) (i, j int, ok bool) { return twoSum(s, target, nop) }

// TraceTwoSum is TwoSum, reporting each pair compared and the pair found.
func TraceTwoSum(s []int, target int, r trace.Recorder) (int, int, bool) {
	return twoSum(s, target, rec{r})
}

func twoSum(s []int, target int, x rec) (int, int, bool) {
	for i, j := 0, len(s)-1; i < j; {
		x.compare(i, j)
		switch sum := s[i] + s[j]; {
		case sum == target:
			x.mark("found", i, j)
			return i, j, true
		case sum < target:
			i++
		default:
			j--
		}
	}
	return 0, 0, false
}

// MaxSubarray returns the largest sum of a non-empty run of consecutive values
// and where it is, s[start:end], with Kadane's algorithm: keep the best sum
// ending here, and start over when it drops below the value alone. It returns
// 0, 0, 0 for an empty s.
//
// Time: O(n). Space: O(1).
func MaxSubarray(s []int) (sum, start, end int) { return kadane(s, nop) }

// TraceMaxSubarray is MaxSubarray, reporting the run being extended (Mark
// "range") and the best run so far (Mark "found").
func TraceMaxSubarray(s []int, r trace.Recorder) (int, int, int) { return kadane(s, rec{r}) }

func kadane(s []int, x rec) (int, int, int) {
	if len(s) == 0 {
		return 0, 0, 0
	}
	best, bs, be := s[0], 0, 1
	cur, cs := s[0], 0
	for i := 1; i < len(s); i++ {
		x.compare(i)
		if cur+s[i] < s[i] {
			cur, cs = s[i], i // starting over beats extending
		} else {
			cur += s[i]
		}
		x.mark("range", cs, i)
		if cur > best {
			best, bs, be = cur, cs, i+1
			x.mark("found", bs, be-1)
		}
	}
	return best, bs, be
}

// WindowMax returns the maximum of every window of k consecutive values, with
// a deque of indexes whose values decrease, so each value enters and leaves it
// once. It returns nil when k is outside 1..len(s).
//
// Time: O(n). Space: O(k).
func WindowMax(s []int, k int) []int { return windowMax(s, k, nop) }

// TraceWindowMax is WindowMax, reporting each window (Mark "range") and its
// maximum (Mark "found").
func TraceWindowMax(s []int, k int, r trace.Recorder) []int { return windowMax(s, k, rec{r}) }

func windowMax(s []int, k int, x rec) []int {
	if k < 1 || k > len(s) {
		return nil
	}
	var out, dq []int // dq holds indexes, values decreasing
	for i, v := range s {
		for len(dq) > 0 && s[dq[len(dq)-1]] <= v {
			x.compare(dq[len(dq)-1], i)
			dq = dq[:len(dq)-1]
		}
		dq = append(dq, i)
		if dq[0] <= i-k {
			dq = dq[1:]
		}
		if i >= k-1 {
			x.mark("range", i-k+1, i)
			x.mark("found", dq[0])
			out = append(out, s[dq[0]])
		}
	}
	return out
}

// PrefixSums returns p with p[i] the sum of s[:i], so the sum of any range
// s[a:b] is p[b] - p[a] in O(1). len(p) is len(s) + 1.
//
// Time: O(n). Space: O(n).
func PrefixSums(s []int) []int { return prefix(s, nop) }

// TracePrefixSums is PrefixSums, reporting each sum written (Set, Indices[0]
// the index into p).
func TracePrefixSums(s []int, r trace.Recorder) []int { return prefix(s, rec{r}) }

func prefix(s []int, x rec) []int {
	p := make([]int, len(s)+1)
	x.set(0, 0)
	for i, v := range s {
		p[i+1] = p[i] + v
		x.set(i+1, p[i+1])
	}
	return p
}

// MergeSorted merges two sorted slices into a new sorted slice, taking the
// smaller front value each step; equal values come from a first.
//
// Time: O(n + m). Space: O(n + m).
func MergeSorted(a, b []int) []int { return mergeSorted(a, b, nop) }

// TraceMergeSorted is MergeSorted, reporting each comparison (Indices[0] in a,
// Indices[1] in b) and each value written to the result (Set).
func TraceMergeSorted(a, b []int, r trace.Recorder) []int { return mergeSorted(a, b, rec{r}) }

func mergeSorted(a, b []int, x rec) []int {
	out := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		x.compare(i, j)
		if a[i] <= b[j] {
			out = append(out, a[i])
			i++
		} else {
			out = append(out, b[j])
			j++
		}
		x.set(len(out)-1, out[len(out)-1])
	}
	for _, v := range append(a[i:], b[j:]...) {
		out = append(out, v)
		x.set(len(out)-1, v)
	}
	return out
}
