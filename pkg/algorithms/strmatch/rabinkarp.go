package strmatch

import "github.com/braswelljr/alGOrithm/pkg/trace"

const (
	rkBase = 256
	rkMod  = 1_000_000_007
)

// RabinKarp finds pattern in text with the Rabin-Karp algorithm (1987):
// compare a rolling hash of each window of the text with the pattern's hash,
// and compare letters only when the hashes agree.
//
// Time: O(n + m) on average, O(n * m) when many hashes collide.
// Space: O(1) beyond the result.
func RabinKarp(text, pattern string) []int { return rabinKarp(text, pattern, nop) }

// TraceRabinKarp is RabinKarp, reporting to r as the package describes.
func TraceRabinKarp(text, pattern string, r trace.Recorder) []int {
	return rabinKarp(text, pattern, rec{r})
}

func rabinKarp(text, pattern string, x rec) []int {
	var out []int
	n, m := len(text), len(pattern)
	if m == 0 || m > n {
		return out
	}
	high := 1 // rkBase^(m-1): the weight of the letter leaving the window
	for i := 1; i < m; i++ {
		high = high * rkBase % rkMod
	}
	var hp, ht int
	for i := 0; i < m; i++ {
		hp = (hp*rkBase + int(pattern[i])) % rkMod
		ht = (ht*rkBase + int(text[i])) % rkMod
	}
	equal := func(s int) bool {
		for j := 0; j < m; j++ {
			x.compare(s+j, j)
			if text[s+j] != pattern[j] {
				return false
			}
		}
		return true
	}
	for s := 0; ; s++ {
		x.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{s, s + m - 1}})
		if ht == hp && equal(s) {
			out = append(out, s)
			x.found(s, s+m-1)
		}
		if s+m >= n {
			return out
		}
		ht = (ht - int(text[s])*high%rkMod + rkMod) % rkMod // drop text[s]
		ht = (ht*rkBase + int(text[s+m])) % rkMod           // add text[s+m]
	}
}
