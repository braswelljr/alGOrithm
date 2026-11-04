// Package arrays is about the structure under almost every other: a run of
// equally sized slots in memory, read and written by index in O(1).
//
// Dynamic is a growable array, like a Go slice or a Python list, built by
// hand so each step can be seen: appending into spare capacity, doubling and
// copying when full, shifting to insert or delete in the middle. The other
// files are the classic array techniques, each traced like the sorts: reverse,
// rotate, two pointers, sliding window, Kadane's maximum subarray, prefix
// sums, and merging two sorted arrays.
//
// Traces use the same events as the sorts (Compare, Swap, Set, Mark), with
// Indices into the array being worked on, so trace.Replay rebuilds the result.
package arrays

import (
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type rec struct{ r trace.Recorder }

func (x rec) compare(i ...int) { x.r.Record(trace.Event{Kind: trace.Compare, Indices: i}) }
func (x rec) swap(i, j int)    { x.r.Record(trace.Event{Kind: trace.Swap, Indices: []int{i, j}}) }
func (x rec) set(i, v int) {
	x.r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{v}})
}
func (x rec) mark(note string, i ...int) {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: i, Note: note})
}

var nop = rec{trace.Nop{}}
