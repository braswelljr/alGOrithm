// Package sorting holds the sorting algorithms, one file each: bubble, cocktail
// shaker, comb, cycle, gnome, heap, insertion, merge, odd-even, pancake, quick,
// selection and Shell sort for any ordered type, and bucket, counting and
// radix sort for integers.
//
// Every algorithm X has two functions sharing one implementation: X sorts, and
// TraceX sorts while reporting each comparison, swap and write to a
// trace.Recorder, which is what the visualizer animates:
//
//	sorted := sorting.Quick([]int{5, 3, 9, 1})
//
//	var log trace.Log
//	sorting.TraceQuick([]int{5, 3, 9, 1}, &log)
//	fmt.Println(log.Count(trace.Compare), log.Count(trace.Swap))
//
// Each function's documentation says whether it sorts in place (the caller's
// slice is modified and returned) or returns a new slice, whether it is
// stable, and what it costs.
package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// rec wraps a Recorder with the events the sorts report.
type rec struct{ r trace.Recorder }

func (x rec) compare(i, j int) {
	x.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{i, j}})
}

func (x rec) swap(i, j int) {
	x.r.Record(trace.Event{Kind: trace.Swap, Indices: []int{i, j}})
}

func (x rec) visit(i int) {
	x.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}})
}

func (x rec) mark(note string, indices ...int) {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: indices, Note: note})
}

// set reports that v was written at i. The value travels with the event when
// T is an integer type, so trace.Replay can rebuild the result; for other
// types only the index is reported.
func set[T cmp.Ordered](x rec, i int, v T) {
	e := trace.Event{Kind: trace.Set, Indices: []int{i}}
	if n, ok := asInt(v); ok {
		e.Values = []int{n}
	}
	x.r.Record(e)
}

func asInt[T cmp.Ordered](v T) (int, bool) {
	switch n := any(v).(type) {
	case int:
		return n, true
	case int8:
		return int(n), true
	case int16:
		return int(n), true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case uint8:
		return int(n), true
	case uint16:
		return int(n), true
	case uint32:
		return int(n), true
	}
	return 0, false
}

// clone returns a copy of s (nil for nil), for the sorts that leave the
// caller's slice alone.
func clone[T any](s []T) []T {
	if s == nil {
		return nil
	}
	return append(make([]T, 0, len(s)), s...)
}

var nop = rec{trace.Nop{}}
