// Package searching holds the search algorithms, one file each: binary,
// exponential, Fibonacci, jump, linear and ternary search for any ordered
// type, interpolation search for integers, and breadth-first search over a
// graph of nodes.
//
// Every index search returns the index of the target, or -1 when it is
// absent. All but Linear need the list sorted in ascending order. Each search X
// has a TraceX that reports the range still searched (Mark "range"), each
// element compared with the target (Compare, one index), and the index found
// (Mark "found"):
//
//	i := searching.Binary([]int{1, 3, 5, 7}, 5) // 2
package searching

import "github.com/braswelljr/alGOrithm/pkg/trace"

type rec struct{ r trace.Recorder }

func (x rec) probe(i int) { x.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{i}}) }

func (x rec) span(lo, hi int) {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: []int{lo, hi}, Note: "range"})
}

func (x rec) found(i int) int {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: []int{i}, Note: "found"})
	return i
}

var nop = rec{trace.Nop{}}
