package text

import "github.com/braswelljr/alGOrithm/pkg/trace"

// LongestCommon returns the longest substring of both a and b as its start in
// a, its start in b and its length; the length is 0 when they share no
// character. It fills one row of the classic table at a time: cell (i, j) is
// the length of the common run ending at a[i-1] and b[j-1].
//
// Time: O(len(a) * len(b)). Space: O(len(b)). Neither text is changed.
func LongestCommon(a, b string) (startA, startB, length int) { return longestCommon(a, b, nop) }

// TraceLongestCommon is LongestCommon, reporting each character pair as
// Compare (Indices[0] in a, Values[0] in b) and the result as Mark "found"
// over a.
func TraceLongestCommon(a, b string, r trace.Recorder) (int, int, int) {
	return longestCommon(a, b, rec{r})
}

func longestCommon(a, b string, x rec) (int, int, int) {
	prev, cur := make([]int, len(b)+1), make([]int, len(b)+1)
	best, endA, endB := 0, 0, 0
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			x.compareTwo(i-1, j-1)
			if a[i-1] == b[j-1] {
				cur[j] = prev[j-1] + 1
				if cur[j] > best {
					best, endA, endB = cur[j], i, j
				}
			} else {
				cur[j] = 0
			}
		}
		prev, cur = cur, prev
	}
	if best > 0 {
		x.found(endA-best, endA)
	}
	return endA - best, endB - best, best
}
