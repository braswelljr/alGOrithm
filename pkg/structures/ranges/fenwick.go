package ranges

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// FenwickTree (a binary indexed tree, Peter Fenwick, 1994) keeps prefix sums
// of an array of ints under point updates: slot i holds the sum of the
// i & -i values ending at i, so both a prefix sum and an update touch
// O(log n) slots. Indexes are 0-based outside, 1-based inside. The zero
// FenwickTree is empty; use NewFenwickTree.
type FenwickTree struct{ tree []int }

// NewFenwickTree builds a tree over values. Time: O(n).
func NewFenwickTree(values []int) *FenwickTree {
	t := &FenwickTree{tree: make([]int, len(values)+1)}
	for i, v := range values {
		t.tree[i+1] += v
		if j := i + 1 + (i+1)&-(i+1); j < len(t.tree) {
			t.tree[j] += t.tree[i+1]
		}
	}
	return t
}

// Len is the number of values. Time: O(1).
func (t *FenwickTree) Len() int { return len(t.tree) - 1 }

// Add adds delta to the value at i and reports whether i was in range.
// Time: O(log n).
func (t *FenwickTree) Add(i, delta int) bool {
	if i < 0 || i >= t.Len() {
		return false
	}
	for j := i + 1; j < len(t.tree); j += j & -j {
		t.tree[j] += delta
	}
	return true
}

// PrefixSum returns the sum of the values at 0..i inclusive; ok is false for
// an i outside the array. Time: O(log n).
func (t *FenwickTree) PrefixSum(i int) (int, bool) {
	if i < 0 || i >= t.Len() {
		return 0, false
	}
	s := 0
	for j := i + 1; j > 0; j -= j & -j {
		s += t.tree[j]
	}
	return s, true
}

// RangeSum returns the sum of the values at l..r inclusive. Time: O(log n).
func (t *FenwickTree) RangeSum(l, r int) (int, bool) {
	if l < 0 || l > r || r >= t.Len() {
		return 0, false
	}
	b, _ := t.PrefixSum(r)
	if l == 0 {
		return b, true
	}
	a, _ := t.PrefixSum(l - 1)
	return b - a, true
}

// Snapshot describes the slots 1..n, each labeled "i: partial sum", with an
// edge from each slot to the one an update climbs to next (i + (i & -i)).
func (t *FenwickTree) Snapshot() trace.Snapshot {
	n := t.Len()
	s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, n)}
	for i := 1; i <= n; i++ {
		s.Nodes[i-1] = trace.Node{ID: i - 1, Label: fmt.Sprintf("%d: %d", i, t.tree[i]), Parent: -1}
		if j := i + i&-i; j <= n {
			s.Edges = append(s.Edges, trace.Edge{From: i - 1, To: j - 1, Label: "up"})
		}
	}
	return s
}
