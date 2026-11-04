package text

import "github.com/braswelljr/alGOrithm/pkg/trace"

// MinimalRotation returns the start k of the lexicographically smallest
// rotation of s, s[k:] + s[:k], by Booth's algorithm (1980): a
// Knuth-Morris-Pratt failure function over s+s that jumps the candidate start
// forward whenever a smaller character turns up. It returns 0 for "".
//
// Time: O(n). Space: O(n). s is not changed.
func MinimalRotation(s string) int { return minimalRotation(s, nop) }

// TraceMinimalRotation is MinimalRotation, reporting each character
// comparison as Compare with both indices taken modulo len(s).
func TraceMinimalRotation(s string, r trace.Recorder) int { return minimalRotation(s, rec{r}) }

func minimalRotation(s string, x rec) int {
	n := len(s)
	if n == 0 {
		return 0
	}
	ss := s + s
	f := make([]int, 2*n)
	for i := range f {
		f[i] = -1
	}
	k := 0
	for j := 1; j < 2*n; j++ {
		i := f[j-k-1]
		for i != -1 && ss[j] != ss[k+i+1] {
			x.compare(j%n, (k+i+1)%n)
			if ss[j] < ss[k+i+1] {
				k = j - i - 1
			}
			i = f[i]
		}
		x.compare(j%n, (k+i+1)%n)
		if ss[j] != ss[k+i+1] { // i == -1
			if ss[j] < ss[k] {
				k = j
			}
			f[j-k] = -1
		} else {
			f[j-k] = i + 1
		}
	}
	x.found(k%n, k%n+1)
	return k % n
}
