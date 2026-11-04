// Package text holds algorithms on whole texts, one file each:
//
//	repeats      LongestRepeated (suffix array + LCP), LongestCommon (two texts),
//	             DistinctSubstrings (suffix automaton)
//	structure    MinimalRotation (Booth), LyndonFactors (Duval), SmallestPeriod
//	search       FuzzySearch (bitap with k errors), Autocomplete (trie),
//	             Suggest (spelling by edit distance), LongestCommonPrefix
//	statistics   WordFrequency, TFIDF, NGrams, Jaccard
//	layout       WrapGreedy, WrapBalanced (minimum raggedness), Justify
//	comparison   Diff (Myers' O(ND) line diff), Soundex
//
// Texts are read as bytes, so the algorithms suit ASCII text; words are
// split on anything that is not a letter, digit or apostrophe. TraceX
// reports each character comparison as Compare with the two text indices
// (for two texts: Indices[0] in the first, Values[0] in the second) and each
// result as Mark "found" over [start, end). Single-text pattern matching
// (KMP, Z, Aho-Corasick, Manacher, suffix arrays) is in package strmatch.
//
//	text.Soundex("Robert") // "R163"
package text

import (
	"errors"
	"strings"
	"unicode"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrInput is wrapped by every error about invalid input.
var ErrInput = errors.New("text: invalid input")

type rec struct{ r trace.Recorder }

func (x rec) compare(i, j int) {
	x.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{i, j}})
}

func (x rec) compareTwo(i, j int) {
	x.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{i}, Values: []int{j}})
}

func (x rec) found(start, end int) {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: []int{start, end}, Note: "found"})
}

var nop = rec{trace.Nop{}}

// Words splits s into lower-case words: runs of letters, digits and
// apostrophes. Time and space O(n).
func Words(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '\''
	})
}
