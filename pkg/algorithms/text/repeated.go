package text

import (
	"sort"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// LongestRepeated returns the longest substring of s that occurs at least
// twice (overlaps allowed) and every index where it starts, or "" and nil
// when no character repeats. It sorts the suffixes by prefix doubling and
// takes the largest common prefix of neighbors (Kasai's LCP array).
//
// Time: O(n log^2 n). Space: O(n). s is not changed.
func LongestRepeated(s string) (string, []int) { return longestRepeated(s, nop) }

// TraceLongestRepeated is LongestRepeated, reporting each neighbor pair of
// sorted suffixes as Compare and each occurrence as Mark "found".
func TraceLongestRepeated(s string, r trace.Recorder) (string, []int) {
	return longestRepeated(s, rec{r})
}

func longestRepeated(s string, x rec) (string, []int) {
	sa := suffixArray(s)
	lcp := lcpArray(s, sa)
	best, at := 0, -1
	for i := 1; i < len(sa); i++ {
		x.compare(sa[i-1], sa[i])
		if lcp[i] > best {
			best, at = lcp[i], sa[i]
		}
	}
	if best == 0 {
		return "", nil
	}
	sub := s[at : at+best]
	var starts []int
	for i := 0; i+best <= len(s); i++ {
		if s[i:i+best] == sub {
			starts = append(starts, i)
			x.found(i, i+best)
		}
	}
	return sub, starts
}

// suffixArray returns the starts of s's suffixes in sorted order, by prefix
// doubling: sort by the first 2^k characters using the ranks of halves.
func suffixArray(s string) []int {
	n := len(s)
	sa, rank, tmp := make([]int, n), make([]int, n), make([]int, n)
	for i := range sa {
		sa[i], rank[i] = i, int(s[i])
	}
	for k := 1; ; k <<= 1 {
		key := func(i int) (int, int) {
			second := -1
			if i+k < n {
				second = rank[i+k]
			}
			return rank[i], second
		}
		less := func(a, b int) bool {
			a1, a2 := key(a)
			b1, b2 := key(b)
			return a1 < b1 || (a1 == b1 && a2 < b2)
		}
		sort.Slice(sa, func(i, j int) bool { return less(sa[i], sa[j]) })
		if n > 0 {
			tmp[sa[0]] = 0
		}
		for i := 1; i < n; i++ {
			tmp[sa[i]] = tmp[sa[i-1]]
			if less(sa[i-1], sa[i]) {
				tmp[sa[i]]++
			}
		}
		copy(rank, tmp)
		if n == 0 || rank[sa[n-1]] == n-1 || k >= n {
			return sa
		}
	}
}

// lcpArray returns lcp[i], the common prefix length of the suffixes at sa[i-1]
// and sa[i] (lcp[0] is 0), by Kasai's algorithm in O(n).
func lcpArray(s string, sa []int) []int {
	n := len(s)
	rank, lcp := make([]int, n), make([]int, n)
	for i, p := range sa {
		rank[p] = i
	}
	h := 0
	for i := 0; i < n; i++ {
		if rank[i] == 0 {
			h = 0
			continue
		}
		j := sa[rank[i]-1]
		for i+h < n && j+h < n && s[i+h] == s[j+h] {
			h++
		}
		lcp[rank[i]] = h
		if h > 0 {
			h--
		}
	}
	return lcp
}

// LongestCommonPrefix returns the longest prefix every word shares, "" for no
// words. Time: O(total length). Space: O(1) beyond the result.
func LongestCommonPrefix(words []string) string {
	if len(words) == 0 {
		return ""
	}
	prefix := words[0]
	for _, w := range words[1:] {
		for !strings.HasPrefix(w, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}
