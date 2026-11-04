// Package spatial holds structures that find points by where they are: a
// KDTree for the nearest point and points in a rectangle, and a QuadTree that
// splits the plane into quarters as points crowd in. Maps, games and
// collision detection are built on them.
package spatial

import "fmt"

// Point is a point with whole-number coordinates.
type Point struct{ X, Y int }

func (p Point) String() string { return fmt.Sprintf("(%d,%d)", p.X, p.Y) }

// Rect is the closed rectangle from Min to Max.
type Rect struct{ Min, Max Point }

// Contains reports whether p is inside r, edges included.
func (r Rect) Contains(p Point) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

// Intersects reports whether r and s share a point.
func (r Rect) Intersects(s Rect) bool {
	return r.Min.X <= s.Max.X && s.Min.X <= r.Max.X && r.Min.Y <= s.Max.Y && s.Min.Y <= r.Max.Y
}

func (r Rect) String() string { return fmt.Sprintf("%v-%v", r.Min, r.Max) }

// Dist2 returns the squared distance between p and q.
func Dist2(p, q Point) int {
	dx, dy := p.X-q.X, p.Y-q.Y
	return dx*dx + dy*dy
}
