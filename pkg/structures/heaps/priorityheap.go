package heaps

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// PriorityHeap is a binary-heap priority queue over any type, ordered by a
// function you give it: less(a, b) reports whether a is more urgent than b.
// With less as < it is a min-heap; with > a max-heap; with a comparison of
// deadlines, a scheduler's run queue. The zero PriorityHeap is not ready; use
// NewPriorityHeap.
type PriorityHeap[T any] struct {
	data []T
	less func(a, b T) bool
}

// NewPriorityHeap returns a heap ordered by less, holding a copy of values.
// Time: O(n).
func NewPriorityHeap[T any](less func(a, b T) bool, values ...T) *PriorityHeap[T] {
	h := &PriorityHeap[T]{data: append([]T(nil), values...), less: less}
	for i := len(h.data)/2 - 1; i >= 0; i-- {
		h.down(i)
	}
	return h
}

// Len is the number of values. Time: O(1).
func (h *PriorityHeap[T]) Len() int { return len(h.data) }

// Push adds v. Time: O(log n).
func (h *PriorityHeap[T]) Push(v T) {
	h.data = append(h.data, v)
	for i := len(h.data) - 1; i > 0; {
		p := (i - 1) / 2
		if !h.less(h.data[i], h.data[p]) {
			break
		}
		h.data[i], h.data[p] = h.data[p], h.data[i]
		i = p
	}
}

// Pop removes and returns the most urgent value; ok is false when empty.
// Time: O(log n).
func (h *PriorityHeap[T]) Pop() (v T, ok bool) {
	if len(h.data) == 0 {
		return v, false
	}
	v = h.data[0]
	last := len(h.data) - 1
	h.data[0] = h.data[last]
	h.data = h.data[:last]
	h.down(0)
	return v, true
}

// Top returns the most urgent value without removing it; ok is false when
// empty. Time: O(1).
func (h *PriorityHeap[T]) Top() (v T, ok bool) {
	if len(h.data) == 0 {
		return v, false
	}
	return h.data[0], true
}

func (h *PriorityHeap[T]) down(i int) {
	for {
		best := i
		for _, c := range [2]int{2*i + 1, 2*i + 2} {
			if c < len(h.data) && h.less(h.data[c], h.data[best]) {
				best = c
			}
		}
		if best == i {
			return
		}
		h.data[i], h.data[best] = h.data[best], h.data[i]
		i = best
	}
}

// Snapshot describes the heap as the tree its slice encodes, node 0 the "root".
func (h *PriorityHeap[T]) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: make([]trace.Node, len(h.data))}
	for i, v := range h.data {
		p := -1
		if i > 0 {
			p = (i - 1) / 2
			s.Edges = append(s.Edges, trace.Edge{From: p, To: i})
		}
		s.Nodes[i] = trace.Node{ID: i, Label: fmt.Sprint(v), Parent: p}
	}
	if len(s.Nodes) > 0 {
		s.Nodes[0].Role = "root"
	}
	return s
}

var _ PriorityQueue[int] = (*PriorityHeap[int])(nil)
