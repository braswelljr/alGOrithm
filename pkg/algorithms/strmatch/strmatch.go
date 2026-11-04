// Package strmatch finds a pattern in a text, one algorithm per file: naive
// search, Knuth-Morris-Pratt and Rabin-Karp. Each returns every index of the
// text where the pattern starts, overlapping matches included; an empty
// pattern matches nowhere. TraceX reports each letter comparison (Compare:
// Indices[0] in the text, Values[0] in the pattern), each window whose hash is
// checked (Visit, Rabin-Karp), and each match (Mark "found" over the matched
// text). The package also has Anagram.
//
//	strmatch.KMP("abracadabra", "abra") // [0 7]
package strmatch

import "github.com/braswelljr/alGOrithm/pkg/trace"

type rec struct{ r trace.Recorder }

func (x rec) compare(textIndex, patternIndex int) {
	x.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{textIndex}, Values: []int{patternIndex}})
}

func (x rec) found(start, end int) {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: []int{start, end}, Note: "found"})
}

var nop = rec{trace.Nop{}}
