package heaps

import (
	"cmp"
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MinHeap is a binary min-heap in a slice: every parent is at most its
// children, so the smallest value is always at the top. The zero MinHeap is
// empty and ready.
type MinHeap[T cmp.Ordered] struct{ data []T }

// NewMinHeap returns a heap holding a copy of values, arranged in O(n). The
// caller's slice is not modified.
func NewMinHeap[T cmp.Ordered](values ...T) *MinHeap[T] {
	h := &MinHeap[T]{data: append([]T(nil), values...)}
	for i := len(h.data)/2 - 1; i >= 0; i-- {
		h.down(i)
	}
	return h
}

// Len is the number of values. Time: O(1).
func (h *MinHeap[T]) Len() int { return len(h.data) }

// Push adds v. Time: O(log n).
func (h *MinHeap[T]) Push(v T) {
	h.data = append(h.data, v)
	for i := len(h.data) - 1; i > 0; {
		p := (i - 1) / 2
		if h.data[i] >= h.data[p] {
			break
		}
		h.data[i], h.data[p] = h.data[p], h.data[i]
		i = p
	}
}

// Pop removes and returns the smallest value; ok is false when empty.
// Time: O(log n).
func (h *MinHeap[T]) Pop() (v T, ok bool) {
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

// Top returns the smallest value without removing it; ok is false when empty.
// Time: O(1).
func (h *MinHeap[T]) Top() (v T, ok bool) {
	if len(h.data) == 0 {
		return v, false
	}
	return h.data[0], true
}

func (h *MinHeap[T]) down(i int) {
	for {
		small := i
		for _, c := range [2]int{2*i + 1, 2*i + 2} {
			if c < len(h.data) && h.data[c] < h.data[small] {
				small = c
			}
		}
		if small == i {
			return
		}
		h.data[i], h.data[small] = h.data[small], h.data[i]
		i = small
	}
}

// Valid reports whether every parent is at most its children. Time: O(n).
func (h *MinHeap[T]) Valid() bool {
	for i := 1; i < len(h.data); i++ {
		if h.data[i] < h.data[(i-1)/2] {
			return false
		}
	}
	return true
}

// Snapshot describes the heap as the binary tree its slice encodes: node i's
// parent is (i-1)/2, node 0 the "root".
func (h *MinHeap[T]) Snapshot() trace.Snapshot {
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
