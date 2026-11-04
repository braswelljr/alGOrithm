package text

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// FuzzySearch returns every index of text where a match of pattern with at
// most k edits (insertions, deletions, substitutions) ends, by the bitap
// (shift-or) algorithm extended to errors by Wu and Manber (1992): one
// machine word per error count holds which prefixes of the pattern match so
// far, and a character updates them with a shift and a mask. The pattern must
// be 1 to 63 bytes long and k from 0 to its length.
//
// Time: O(n * k). Space: O(k + alphabet). Neither string is changed.
func FuzzySearch(text, pattern string, k int) ([]int, error) {
	return fuzzySearch(text, pattern, k, nop)
}

// TraceFuzzySearch is FuzzySearch, reporting each text character read as
// Compare and each match as Mark "found" over the len(pattern) characters
// that end there (or fewer at the start of the text).
func TraceFuzzySearch(text, pattern string, k int, r trace.Recorder) ([]int, error) {
	return fuzzySearch(text, pattern, k, rec{r})
}

func fuzzySearch(text, pattern string, k int, x rec) ([]int, error) {
	m := len(pattern)
	if m == 0 || m > 63 {
		return nil, fmt.Errorf("%w: the pattern must be 1 to 63 bytes, not %d", ErrInput, m)
	}
	if k < 0 || k > m {
		return nil, fmt.Errorf("%w: errors must be 0 to %d, not %d", ErrInput, m, k)
	}
	var mask [256]uint64
	for i := 0; i < m; i++ {
		mask[pattern[i]] |= 1 << i
	}
	// r[d] bit i: pattern[:i+1] matches a suffix of the text read so far with
	// at most d edits. Before any text, d deletions match d characters.
	r := make([]uint64, k+1)
	for d := range r {
		r[d] = 1<<d - 1
	}
	done := uint64(1) << (m - 1)
	ends := []int{}
	for j := 0; j < len(text); j++ {
		x.compare(j, j)
		b := mask[text[j]]
		old := r[0]
		r[0] = (r[0]<<1 | 1) & b
		for d := 1; d <= k; d++ {
			prev := r[d]
			r[d] = (r[d]<<1|1)&b | // match
				(old<<1 | 1) | // substitution
				old | // an extra text character
				(r[d-1]<<1 | 1) // a skipped pattern character
			old = prev
		}
		if r[k]&done != 0 {
			ends = append(ends, j)
			x.found(max(0, j-m+1), j+1)
		}
	}
	return ends, nil
}
