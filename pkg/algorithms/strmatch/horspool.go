package strmatch

import "github.com/braswelljr/alGOrithm/pkg/trace"

// BoyerMooreHorspool finds pattern in text with Horspool's simplification of
// Boyer-Moore (1980): compare each window from its right end, and on a
// mismatch shift by how far the window's last letter is from its last place in
// the pattern, often skipping many positions at once.
//
// Time: O(n / m) at best, O(n * m) at worst; usually sublinear in practice.
// Space: O(256) for the shift table.
func BoyerMooreHorspool(text, pattern string) []int { return horspool(text, pattern, nop) }

// TraceBoyerMooreHorspool is BoyerMooreHorspool, reporting each window tried
// (Visit over the window), each letter compared and each match.
func TraceBoyerMooreHorspool(text, pattern string, r trace.Recorder) []int {
	return horspool(text, pattern, rec{r})
}

func horspool(text, pattern string, x rec) []int {
	var out []int
	n, m := len(text), len(pattern)
	if m == 0 || m > n {
		return out
	}
	var shift [256]int
	for c := range shift {
		shift[c] = m
	}
	for i := 0; i < m-1; i++ {
		shift[pattern[i]] = m - 1 - i
	}
	for s := 0; s+m <= n; s += shift[text[s+m-1]] {
		x.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{s, s + m - 1}})
		j := m - 1
		for j >= 0 {
			x.compare(s+j, j)
			if text[s+j] != pattern[j] {
				break
			}
			j--
		}
		if j < 0 {
			out = append(out, s)
			x.found(s, s+m-1)
		}
	}
	return out
}
