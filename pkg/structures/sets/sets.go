// Package sets holds collections without duplicates and the operations on
// them, one file each: Bitset (a set of small integers, one bit each),
// Multiset (a bag: values with counts), OrderedSet (a sorted set with rank and
// range queries), SparseSet (small integers with O(1) clear and iteration),
// Roaring (a compressed bitmap of 32-bit integers), and the set operations
// on sorted slices (union, intersection, difference, symmetric difference)
// done with two pointers, plus the power set, the Cartesian product and
// Jaccard similarity.
//
// The sorted-slice operations have TraceX forms that report each comparison
// (Compare: Indices[0] into a, Indices[1] into b) and each value written to
// the result (Set: Indices[0] the result index, Values[0] the value), so they
// animate like the merge step of merge sort:
//
//	sets.Union([]int{1, 3, 5}, []int{2, 3, 4}) // [1 2 3 4 5]
package sets

import "github.com/braswelljr/alGOrithm/pkg/trace"

type rec struct{ r trace.Recorder }

func (x rec) compare(i, j int) {
	x.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{i, j}})
}

func (x rec) set(i, v int) {
	x.r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{v}})
}

var nop = rec{trace.Nop{}}
