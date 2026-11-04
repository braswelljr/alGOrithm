package strmatch

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Naive finds pattern in text the simplest way: try every starting position
// and compare letter by letter.
//
// Time: O(n * m) for a text of n bytes and a pattern of m.
// Space: O(1) beyond the result.
func Naive(text, pattern string) []int { return naive(text, pattern, nop) }

// TraceNaive is Naive, reporting to r as the package describes.
func TraceNaive(text, pattern string, r trace.Recorder) []int { return naive(text, pattern, rec{r}) }

func naive(text, pattern string, x rec) []int {
	var out []int
	n, m := len(text), len(pattern)
	if m == 0 {
		return out
	}
	for s := 0; s+m <= n; s++ {
		j := 0
		for j < m {
			x.compare(s+j, j)
			if text[s+j] != pattern[j] {
				break
			}
			j++
		}
		if j == m {
			out = append(out, s)
			x.found(s, s+m-1)
		}
	}
	return out
}
