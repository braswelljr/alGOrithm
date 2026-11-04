package strmatch

import "github.com/braswelljr/alGOrithm/pkg/trace"

// ZArray returns z with z[i] the length of the longest substring starting at
// i that is also a prefix of s (z[0] is len(s)), computed in one pass by
// reusing the rightmost match found so far, the Z-box.
//
// Time: O(n). Space: O(n).
func ZArray(s string) []int { return zArray(s, nil) }

func zArray(s string, compare func(i, j int)) []int {
	n := len(s)
	z := make([]int, n)
	if n == 0 {
		return z
	}
	z[0] = n
	for i, l, r := 1, 0, 0; i < n; i++ {
		if i < r {
			z[i] = min(r-i, z[i-l]) // inside the Z-box: start from the mirrored value
		}
		for i+z[i] < n {
			if compare != nil {
				compare(i+z[i], z[i])
			}
			if s[z[i]] != s[i+z[i]] {
				break
			}
			z[i]++
		}
		if i+z[i] > r {
			l, r = i, i+z[i]
		}
	}
	return z
}

// ZSearch finds pattern in text with the Z-algorithm: compute the Z array of
// pattern + a separator + text; wherever a value reaches len(pattern), the
// pattern starts.
//
// Time: O(n + m). Space: O(n + m).
func ZSearch(text, pattern string) []int { return zSearch(text, pattern, nop) }

// TraceZSearch is ZSearch, reporting each letter compared in the text (with
// its place in the pattern) and each match.
func TraceZSearch(text, pattern string, r trace.Recorder) []int {
	return zSearch(text, pattern, rec{r})
}

func zSearch(text, pattern string, x rec) []int {
	var out []int
	m := len(pattern)
	if m == 0 || m > len(text) {
		return out
	}
	s := pattern + "\x00" + text
	z := zArray(s, func(i, j int) {
		if i > m && j < m {
			x.compare(i-m-1, j)
		}
	})
	for i := m + 1; i < len(s); i++ {
		if z[i] >= m {
			out = append(out, i-m-1)
			x.found(i-m-1, i-2)
		}
	}
	return out
}
