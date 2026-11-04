package spatial

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// maxQuadDepth stops a QuadTree splitting forever on points stacked together.
const maxQuadDepth = 16

// QuadTree divides a square region into four quarters whenever one holds more
// than capacity points (Raphael Finkel and Jon Bentley, 1974), so crowded
// areas get small cells and empty ones stay whole. Query visits only the
// cells that overlap the rectangle asked about, which is how a game finds what
// is near the player without checking every object.
//
// Time: O(depth) insert; query proportional to the cells it overlaps plus
// the points found.
// Space: O(n).
type QuadTree struct {
	bounds   Rect
	capacity int
	depth    int
	points   []Point
	kids     *[4]*QuadTree // nil until split
}

// NewQuadTree returns an empty tree over bounds whose cells hold capacity
// points (at least 1) before splitting.
func NewQuadTree(bounds Rect, capacity int) *QuadTree {
	return &QuadTree{bounds: bounds, capacity: max(capacity, 1)}
}

// Insert adds p and reports whether it is inside the bounds and not there already.
func (q *QuadTree) Insert(p Point) bool {
	if !q.bounds.Contains(p) || q.Contains(p) {
		return false
	}
	q.insert(p)
	return true
}

func (q *QuadTree) insert(p Point) {
	if q.kids != nil {
		q.kids[q.quarter(p)].insert(p)
		return
	}
	q.points = append(q.points, p)
	w, h := q.bounds.Max.X-q.bounds.Min.X, q.bounds.Max.Y-q.bounds.Min.Y
	if len(q.points) > q.capacity && q.depth < maxQuadDepth && (w > 0 || h > 0) {
		q.split()
	}
}

func (q *QuadTree) split() {
	mx, my := (q.bounds.Min.X+q.bounds.Max.X)/2, (q.bounds.Min.Y+q.bounds.Max.Y)/2
	b := q.bounds
	quads := [4]Rect{
		{b.Min, Point{mx, my}},                       // south-west
		{Point{mx + 1, b.Min.Y}, Point{b.Max.X, my}}, // south-east
		{Point{b.Min.X, my + 1}, Point{mx, b.Max.Y}}, // north-west
		{Point{mx + 1, my + 1}, b.Max},               // north-east
	}
	q.kids = &[4]*QuadTree{}
	for i, r := range quads {
		q.kids[i] = &QuadTree{bounds: r, capacity: q.capacity, depth: q.depth + 1}
	}
	pts := q.points
	q.points = nil
	for _, p := range pts {
		q.kids[q.quarter(p)].insert(p)
	}
}

func (q *QuadTree) quarter(p Point) int {
	mx, my := (q.bounds.Min.X+q.bounds.Max.X)/2, (q.bounds.Min.Y+q.bounds.Max.Y)/2
	i := 0
	if p.X > mx {
		i++
	}
	if p.Y > my {
		i += 2
	}
	return i
}

// Contains reports whether p is in the tree.
func (q *QuadTree) Contains(p Point) bool {
	if !q.bounds.Contains(p) {
		return false
	}
	if q.kids != nil {
		return q.kids[q.quarter(p)].Contains(p)
	}
	for _, x := range q.points {
		if x == p {
			return true
		}
	}
	return false
}

// Query returns the points inside r and how many cells it looked in.
func (q *QuadTree) Query(r Rect) (found []Point, cells int) {
	var walk func(n *QuadTree)
	walk = func(n *QuadTree) {
		if !n.bounds.Intersects(r) {
			return
		}
		cells++
		if n.kids != nil {
			for _, k := range n.kids {
				walk(k)
			}
			return
		}
		for _, p := range n.points {
			if r.Contains(p) {
				found = append(found, p)
			}
		}
	}
	walk(q)
	return found, cells
}

// Len returns the number of points.
func (q *QuadTree) Len() int {
	if q.kids == nil {
		return len(q.points)
	}
	n := 0
	for _, k := range q.kids {
		n += k.Len()
	}
	return n
}

// Snapshot describes the cells as a tree; a leaf lists its points.
func (q *QuadTree) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	names := [4]string{"SW", "SE", "NW", "NE"}
	var walk func(n *QuadTree, parent int, side string)
	walk = func(n *QuadTree, parent int, side string) {
		id := len(s.Nodes)
		label := n.bounds.String()
		if n.kids == nil {
			label += fmt.Sprintf(" %v", n.points)
		}
		nd := trace.Node{ID: id, Label: label, Parent: parent}
		if parent < 0 {
			nd.Role = "root"
		} else {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id, Label: side})
		}
		s.Nodes = append(s.Nodes, nd)
		if n.kids != nil {
			for i, k := range n.kids {
				walk(k, id, names[i])
			}
		}
	}
	walk(q, -1, "")
	return s
}
