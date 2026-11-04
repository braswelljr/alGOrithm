package strmatch

import "github.com/braswelljr/alGOrithm/pkg/trace"

// KMP finds pattern in text with the Knuth-Morris-Pratt algorithm (1977):
// after a mismatch it never re-reads text it has already matched, because a
// table built from the pattern (KMPTable) says how much of the match can be
// kept.
//
// Time: O(n + m) for a text of n bytes and a pattern of m.
// Space: O(m) for the table.
func KMP(text, pattern string) []int { return kmp(text, pattern, nop) }

// TraceKMP is KMP, reporting to r as the package describes.
func TraceKMP(text, pattern string, r trace.Recorder) []int { return kmp(text, pattern, rec{r}) }

// KMPTable returns the failure table: table[i] is the length of the longest
// proper prefix of pattern[:i+1] that is also its suffix.
//
// Time: O(m). Space: O(m).
func KMPTable(pattern string) []int {
	t := make([]int, len(pattern))
	k := 0
	for i := 1; i < len(pattern); i++ {
		for k > 0 && pattern[i] != pattern[k] {
			k = t[k-1]
		}
		if pattern[i] == pattern[k] {
			k++
		}
		t[i] = k
	}
	return t
}

func kmp(text, pattern string, x rec) []int {
	var out []int
	m := len(pattern)
	if m == 0 {
		return out
	}
	t := KMPTable(pattern)
	j := 0
	for i := 0; i < len(text); i++ {
		for {
			x.compare(i, j)
			if text[i] == pattern[j] || j == 0 {
				break
			}
			j = t[j-1] // keep the longest part of the match that still fits
		}
		if text[i] == pattern[j] {
			j++
		}
		if j == m {
			out = append(out, i-m+1)
			x.found(i-m+1, i)
			j = t[j-1]
		}
	}
	return out
}
