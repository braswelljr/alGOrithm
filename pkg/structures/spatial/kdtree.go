package spatial

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type kdNode struct {
	p           Point
	left, right *kdNode
}

// KDTree is a binary tree of points that splits on x at even depths and on y
// at odd ones (Jon Bentley, 1975): every point left of a node is smaller in
// that node's axis. Nearest walks down toward the query, then back up,
// visiting the far side of a split only when the circle around the best
// point so far crosses the splitting line, which on spread-out points skips
// most of the tree.
//
// Inserts are not rebalanced, so the height h is O(log n) for points in random
// order; NewKDTree builds a balanced tree by splitting at medians.
//
// Time: O(h) insert, about O(log n) nearest on balanced, spread-out points
// (O(n) worst), O(sqrt n + k) range for k points found on a balanced tree.
// Space: O(n).
type KDTree struct {
	root *kdNode
	n    int
}

// NewKDTree builds a balanced tree over points (duplicates are kept once).
func NewKDTree(points ...Point) *KDTree {
	seen := map[Point]bool{}
	var ps []Point
	for _, p := range points {
		if !seen[p] {
			seen[p] = true
			ps = append(ps, p)
		}
	}
	var build func(ps []Point, depth int) *kdNode
	build = func(ps []Point, depth int) *kdNode {
		if len(ps) == 0 {
			return nil
		}
		sortByAxis(ps, depth%2)
		m := len(ps) / 2
		return &kdNode{p: ps[m], left: build(ps[:m], depth+1), right: build(ps[m+1:], depth+1)}
	}
	return &KDTree{root: build(ps, 0), n: len(ps)}
}

func axis(p Point, a int) int {
	if a == 0 {
		return p.X
	}
	return p.Y
}

// sortByAxis sorts ps by axis a, then by the other axis, so equal
// coordinates split the same way every time (insertion sort: KISS, and the
// slices here are small).
func sortByAxis(ps []Point, a int) {
	less := func(p, q Point) bool {
		if axis(p, a) != axis(q, a) {
			return axis(p, a) < axis(q, a)
		}
		return axis(p, 1-a) < axis(q, 1-a)
	}
	for i := 1; i < len(ps); i++ {
		for j := i; j > 0 && less(ps[j], ps[j-1]); j-- {
			ps[j], ps[j-1] = ps[j-1], ps[j]
		}
	}
}

// Insert adds p and reports whether it was not there already.
func (t *KDTree) Insert(p Point) bool {
	link := &t.root
	for depth := 0; *link != nil; depth++ {
		n := *link
		if n.p == p {
			return false
		}
		if axis(p, depth%2) < axis(n.p, depth%2) {
			link = &n.left
		} else {
			link = &n.right
		}
	}
	*link = &kdNode{p: p}
	t.n++
	return true
}

// Nearest returns the point closest to q, its squared distance, and how
// many nodes the search visited; ok is false when the tree is empty.
func (t *KDTree) Nearest(q Point) (best Point, dist2, visited int, ok bool) {
	if t.root == nil {
		return Point{}, 0, 0, false
	}
	dist2 = -1
	var walk func(n *kdNode, depth int)
	walk = func(n *kdNode, depth int) {
		if n == nil {
			return
		}
		visited++
		if d := Dist2(n.p, q); dist2 < 0 || d < dist2 {
			best, dist2 = n.p, d
		}
		a := depth % 2
		diff := axis(q, a) - axis(n.p, a)
		near, far := n.left, n.right
		if diff >= 0 {
			near, far = n.right, n.left
		}
		walk(near, depth+1)
		if diff*diff < dist2 { // the best circle crosses the split: look across
			walk(far, depth+1)
		}
	}
	walk(t.root, 0)
	return best, dist2, visited, true
}

// InRect returns every point inside r, in tree order.
func (t *KDTree) InRect(r Rect) []Point {
	var out []Point
	var walk func(n *kdNode, depth int)
	walk = func(n *kdNode, depth int) {
		if n == nil {
			return
		}
		if r.Contains(n.p) {
			out = append(out, n.p)
		}
		a := depth % 2
		if axis(r.Min, a) <= axis(n.p, a) { // the build can put ties on the left
			walk(n.left, depth+1)
		}
		if axis(r.Max, a) >= axis(n.p, a) {
			walk(n.right, depth+1)
		}
	}
	walk(t.root, 0)
	return out
}

// Len returns the number of points.
func (t *KDTree) Len() int { return t.n }

// Snapshot describes the tree; each label is the point and the axis it splits.
func (t *KDTree) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	var walk func(n *kdNode, parent, depth int, side string)
	walk = func(n *kdNode, parent, depth int, side string) {
		if n == nil {
			return
		}
		id := len(s.Nodes)
		nd := trace.Node{ID: id, Label: fmt.Sprintf("%v %s", n.p, [2]string{"x", "y"}[depth%2]), Parent: parent}
		if parent < 0 {
			nd.Role = "root"
		} else {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id, Label: side})
		}
		s.Nodes = append(s.Nodes, nd)
		walk(n.left, id, depth+1, "left")
		walk(n.right, id, depth+1, "right")
	}
	walk(t.root, -1, 0, "")
	return s
}
