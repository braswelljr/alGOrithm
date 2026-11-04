package text

import "github.com/braswelljr/alGOrithm/pkg/trace"

// SmallestPeriod returns the smallest p > 0 with s[i] == s[i+p] for every i,
// from the prefix function (the longest proper border of s is n - p), and
// whether s is a whole number of copies of s[:p] ("abab" is, "abca" is not).
// It returns 0 and false for "".
//
// Time: O(n). Space: O(n). s is not changed.
func SmallestPeriod(s string) (period int, whole bool) { return smallestPeriod(s, nop) }

// TraceSmallestPeriod is SmallestPeriod, reporting each comparison of the
// prefix function and the period as Mark "found" over s[:p].
func TraceSmallestPeriod(s string, r trace.Recorder) (int, bool) { return smallestPeriod(s, rec{r}) }

func smallestPeriod(s string, x rec) (int, bool) {
	n := len(s)
	if n == 0 {
		return 0, false
	}
	pi := make([]int, n)
	for i := 1; i < n; i++ {
		k := pi[i-1]
		for k > 0 && s[i] != s[k] {
			x.compare(i, k)
			k = pi[k-1]
		}
		x.compare(i, k)
		if s[i] == s[k] {
			k++
		}
		pi[i] = k
	}
	p := n - pi[n-1]
	x.found(0, p)
	return p, n%p == 0
}
