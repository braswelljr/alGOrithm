package strmatch

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// SuffixArray returns the starting positions of all suffixes of s in sorted
// order, built by prefix doubling: sort by the first 1, 2, 4, ... letters,
// each round ranking pairs of the previous round's ranks; and the LCP array
// (Kasai et al., 2001), lcp[i] the length of the common prefix of the suffixes
// at sa[i-1] and sa[i] (lcp[0] is 0). A suffix array answers "where does this
// pattern occur?" with a binary search.
//
// Time: O(n log^2 n) for the array, O(n) for the LCP. Space: O(n).
func SuffixArray(s string) (sa, lcp []int) { return suffixArray(s, trace.Nop{}) }

// TraceSuffixArray is SuffixArray, reporting the order after each doubling
// round (Set, Indices[0] the rank position, Values[0] the suffix start) and
// the doubled length (Mark "range", Values[0] the length compared).
func TraceSuffixArray(s string, r trace.Recorder) ([]int, []int) { return suffixArray(s, r) }

func suffixArray(s string, r trace.Recorder) ([]int, []int) {
	n := len(s)
	sa := make([]int, n)
	rank := make([]int, n)
	for i := range sa {
		sa[i], rank[i] = i, int(s[i])
	}
	tmp := make([]int, n)
	for k := 1; n > 0; k *= 2 {
		key := func(i int) (int, int) {
			second := -1
			if i+k < n {
				second = rank[i+k]
			}
			return rank[i], second
		}
		sort.Slice(sa, func(a, b int) bool {
			a1, a2 := key(sa[a])
			b1, b2 := key(sa[b])
			if a1 != b1 {
				return a1 < b1
			}
			return a2 < b2
		})
		r.Record(trace.Event{Kind: trace.Mark, Values: []int{k}, Note: "range"})
		for i, v := range sa {
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{v}})
		}
		tmp[sa[0]] = 0
		for i := 1; i < n; i++ {
			p1, p2 := key(sa[i-1])
			c1, c2 := key(sa[i])
			tmp[sa[i]] = tmp[sa[i-1]]
			if p1 != c1 || p2 != c2 {
				tmp[sa[i]]++
			}
		}
		copy(rank, tmp)
		if rank[sa[n-1]] == n-1 || k >= n {
			break // every suffix has its own rank
		}
	}
	lcp := make([]int, n)
	inv := make([]int, n)
	for i, v := range sa {
		inv[v] = i
	}
	h := 0
	for i := 0; i < n; i++ {
		if inv[i] == 0 {
			h = 0
			continue
		}
		j := sa[inv[i]-1]
		for i+h < n && j+h < n && s[i+h] == s[j+h] {
			h++
		}
		lcp[inv[i]] = h
		if h > 0 {
			h--
		}
	}
	return sa, lcp
}
