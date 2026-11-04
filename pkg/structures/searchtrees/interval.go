package searchtrees

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Interval is the closed range [Lo, Hi] of whole numbers.
type Interval struct{ Lo, Hi int }

// Overlaps reports whether i and j share at least one point.
func (i Interval) Overlaps(j Interval) bool { return i.Lo <= j.Hi && j.Lo <= i.Hi }

func (i Interval) String() string { return fmt.Sprintf("[%d,%d]", i.Lo, i.Hi) }

type intervalNode struct {
	iv          Interval
	max         int // the largest Hi in this subtree
	left, right *intervalNode
}

// IntervalTree holds intervals and finds every one that overlaps a query
// (calendars, genome ranges, "which bookings clash with this one"). It is a
// binary search tree ordered by Lo in which each node also keeps the largest
// Hi below it, so a search skips every subtree that ends before the query
// starts (Cormen et al., Introduction to Algorithms, 14.3).
//
// This tree is not rebalanced, to keep the augmentation easy to read: its
// height h is O(log n) for intervals inserted in random order and O(n) for
// sorted ones.
//
// Time: O(h) insert and remove; O(h + k) to report k overlaps.
// Space: O(n).
type IntervalTree struct {
	root *intervalNode
	n    int
}

// NewIntervalTree returns an empty interval tree.
func NewIntervalTree() *IntervalTree { return &IntervalTree{} }

func less(a, b Interval) bool { return a.Lo < b.Lo || (a.Lo == b.Lo && a.Hi < b.Hi) }

func (n *intervalNode) fix() {
	n.max = n.iv.Hi
	if n.left != nil && n.left.max > n.max {
		n.max = n.left.max
	}
	if n.right != nil && n.right.max > n.max {
		n.max = n.right.max
	}
}

// Insert adds iv and reports whether it was not there already. An interval
// with Lo > Hi is refused.
func (t *IntervalTree) Insert(iv Interval) bool {
	if iv.Lo > iv.Hi {
		return false
	}
	added := false
	var ins func(n *intervalNode) *intervalNode
	ins = func(n *intervalNode) *intervalNode {
		switch {
		case n == nil:
			added = true
			return &intervalNode{iv: iv, max: iv.Hi}
		case less(iv, n.iv):
			n.left = ins(n.left)
		case less(n.iv, iv):
			n.right = ins(n.right)
		}
		n.fix()
		return n
	}
	t.root = ins(t.root)
	if added {
		t.n++
	}
	return added
}

// Remove deletes iv and reports whether it was there.
func (t *IntervalTree) Remove(iv Interval) bool {
	removed := false
	var del func(n *intervalNode, iv Interval) *intervalNode
	del = func(n *intervalNode, iv Interval) *intervalNode {
		switch {
		case n == nil:
			return nil
		case less(iv, n.iv):
			n.left = del(n.left, iv)
		case less(n.iv, iv):
			n.right = del(n.right, iv)
		default:
			removed = true
			if n.left == nil {
				return n.right
			}
			if n.right == nil {
				return n.left
			}
			m := n.right
			for m.left != nil {
				m = m.left
			}
			n.iv = m.iv
			n.right = del(n.right, m.iv)
		}
		n.fix()
		return n
	}
	t.root = del(t.root, iv)
	if removed {
		t.n--
	}
	return removed
}

// Overlapping returns every interval that overlaps q, ordered by Lo.
func (t *IntervalTree) Overlapping(q Interval) []Interval {
	var out []Interval
	var walk func(n *intervalNode)
	walk = func(n *intervalNode) {
		if n == nil || n.max < q.Lo { // nothing below ends late enough
			return
		}
		walk(n.left)
		if n.iv.Overlaps(q) {
			out = append(out, n.iv)
		}
		if n.iv.Lo <= q.Hi { // to the right every Lo is at least n.iv.Lo
			walk(n.right)
		}
	}
	walk(t.root)
	return out
}

// Len returns the number of intervals.
func (t *IntervalTree) Len() int { return t.n }

// Intervals returns every interval, ordered by Lo, then Hi (an in-order walk).
func (t *IntervalTree) Intervals() []Interval {
	return t.Overlapping(Interval{Lo: -1 << 62, Hi: 1 << 62})
}

// Snapshot describes the tree; each label is the interval and its subtree's max.
func (t *IntervalTree) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	var walk func(n *intervalNode, parent int, side string)
	walk = func(n *intervalNode, parent int, side string) {
		if n == nil {
			return
		}
		id := len(s.Nodes)
		nd := trace.Node{ID: id, Label: fmt.Sprintf("%v max %d", n.iv, n.max), Parent: parent}
		if parent < 0 {
			nd.Role = "root"
		} else {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id, Label: side})
		}
		s.Nodes = append(s.Nodes, nd)
		walk(n.left, id, "left")
		walk(n.right, id, "right")
	}
	walk(t.root, -1, "")
	return s
}
