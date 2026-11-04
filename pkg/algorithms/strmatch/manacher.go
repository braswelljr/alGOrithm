package strmatch

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Manacher finds the longest palindromic substring of s in linear time
// (Glenn Manacher, 1975): put a separator between letters so even and odd
// palindromes look alike, then grow each center's radius starting from its
// mirror inside the rightmost palindrome found so far. It returns the start
// and length of the first longest one, and the radii (one per position of the
// separated string "#a#b#").
//
// Time: O(n). Space: O(n).
func Manacher(s string) (start, length int, radius []int) { return manacher(s, nop) }

// TraceManacher is Manacher, reporting each letter comparison as the radius
// grows (Compare, Indices the two text positions), each radius written (Set,
// Indices[0] the position in the separated string), and the answer (Mark
// "found" over the text).
func TraceManacher(s string, r trace.Recorder) (int, int, []int) { return manacher(s, rec{r}) }

func manacher(s string, x rec) (int, int, []int) {
	if len(s) == 0 {
		return 0, 0, nil
	}
	t := make([]byte, 2*len(s)+1) // "#a#b#"
	for i := range t {
		t[i] = '#'
		if i%2 == 1 {
			t[i] = s[i/2]
		}
	}
	p := make([]int, len(t))
	center, right := 0, 0
	for i := range t {
		if i < right {
			p[i] = min(right-i, p[2*center-i])
		}
		for i-p[i]-1 >= 0 && i+p[i]+1 < len(t) {
			a, b := i-p[i]-1, i+p[i]+1
			if a%2 == 1 { // a letter, not a separator: a real comparison
				x.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{a / 2, b / 2}})
			}
			if t[a] != t[b] {
				break
			}
			p[i]++
		}
		x.r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{p[i]}})
		if i+p[i] > right {
			center, right = i, i+p[i]
		}
	}
	best := 0
	for i := range p {
		if p[i] > p[best] {
			best = i
		}
	}
	start, length := (best-p[best])/2, p[best]
	x.found(start, start+length-1)
	return start, length, p
}
