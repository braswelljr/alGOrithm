package ranges

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// SegmentTree answers range sums and range minimums over an array of ints,
// with updates: each node holds the sum and minimum of a range, its children
// the two halves, so any range is covered by O(log n) nodes. The zero
// SegmentTree is empty; use NewSegmentTree.
type SegmentTree struct {
	n        int
	sum, min []int // node 1 is the root; node i's children are 2i and 2i+1
}

// NewSegmentTree builds a tree over a copy of values. Time: O(n).
func NewSegmentTree(values []int) *SegmentTree {
	t := &SegmentTree{n: len(values), sum: make([]int, 4*max(1, len(values))), min: make([]int, 4*max(1, len(values)))}
	if t.n > 0 {
		t.build(values, 1, 0, t.n-1)
	}
	return t
}

func (t *SegmentTree) build(v []int, node, l, r int) {
	if l == r {
		t.sum[node], t.min[node] = v[l], v[l]
		return
	}
	m := (l + r) / 2
	t.build(v, 2*node, l, m)
	t.build(v, 2*node+1, m+1, r)
	t.pull(node)
}

func (t *SegmentTree) pull(node int) {
	t.sum[node] = t.sum[2*node] + t.sum[2*node+1]
	t.min[node] = min(t.min[2*node], t.min[2*node+1])
}

// Len is the number of values. Time: O(1).
func (t *SegmentTree) Len() int { return t.n }

// Update sets the value at i and reports whether i was in range. Time: O(log n).
func (t *SegmentTree) Update(i, v int) bool {
	if i < 0 || i >= t.n {
		return false
	}
	var set func(node, l, r int)
	set = func(node, l, r int) {
		if l == r {
			t.sum[node], t.min[node] = v, v
			return
		}
		if m := (l + r) / 2; i <= m {
			set(2*node, l, m)
		} else {
			set(2*node+1, m+1, r)
		}
		t.pull(node)
	}
	set(1, 0, t.n-1)
	return true
}

// query folds the nodes covering [ql, qr] with f.
func (t *SegmentTree) query(ql, qr int, of func(node int) int, f func(a, b int) int) (int, bool) {
	if ql < 0 || qr >= t.n || ql > qr {
		return 0, false
	}
	var q func(node, l, r int) (int, bool)
	q = func(node, l, r int) (int, bool) {
		if qr < l || r < ql {
			return 0, false
		}
		if ql <= l && r <= qr {
			return of(node), true
		}
		m := (l + r) / 2
		a, oka := q(2*node, l, m)
		b, okb := q(2*node+1, m+1, r)
		switch {
		case oka && okb:
			return f(a, b), true
		case oka:
			return a, true
		}
		return b, okb
	}
	return q(1, 0, t.n-1)
}

// Sum returns the sum of the values at l..r inclusive; ok is false for a
// range outside the array. Time: O(log n).
func (t *SegmentTree) Sum(l, r int) (int, bool) {
	return t.query(l, r, func(n int) int { return t.sum[n] }, func(a, b int) int { return a + b })
}

// Min returns the smallest value at l..r inclusive; ok is false for a range
// outside the array. Time: O(log n).
func (t *SegmentTree) Min(l, r int) (int, bool) {
	return t.query(l, r, func(n int) int { return t.min[n] }, func(a, b int) int { return min(a, b) })
}

// Snapshot describes the tree, each node labeled "[l,r]: sum".
func (t *SegmentTree) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	if t.n == 0 {
		return s
	}
	var walk func(node, l, r, parent int)
	walk = func(node, l, r, parent int) {
		id := len(s.Nodes)
		nd := trace.Node{ID: id, Label: fmt.Sprintf("[%d,%d]: %d", l, r, t.sum[node]), Parent: parent}
		if parent < 0 {
			nd.Role = "root"
		} else {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id})
		}
		if l == r {
			nd.Role = "leaf"
		}
		s.Nodes = append(s.Nodes, nd)
		if l < r {
			m := (l + r) / 2
			walk(2*node, l, m, id)
			walk(2*node+1, m+1, r, id)
		}
	}
	walk(1, 0, t.n-1, -1)
	return s
}
