package geometry

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ConvexHullGraham returns the convex hull of pts with the Graham scan
// (Ronald Graham, 1972): sort the points by angle around the lowest one, then
// walk them keeping a stack that only turns left. The result is point indexes
// in counterclockwise order, starting from the lowest (then leftmost) point,
// without collinear points on the edges; fewer than three distinct points
// give what there is.
//
// Time: O(n log n). Space: O(n).
func ConvexHullGraham(pts []Point) []int { return graham(pts, nop) }

// TraceConvexHullGraham is ConvexHullGraham, reporting each turn test
// (Compare of three points) and the hull after each step (Mark "hull").
func TraceConvexHullGraham(pts []Point, r trace.Recorder) []int { return graham(pts, rec{r}) }

func graham(pts []Point, x rec) []int {
	idx := distinct(pts)
	if len(idx) < 3 {
		return idx
	}
	low := 0
	for i, p := range idx {
		q := pts[idx[low]]
		if pts[p].Y < q.Y || (pts[p].Y == q.Y && pts[p].X < q.X) {
			low = i
		}
	}
	idx[0], idx[low] = idx[low], idx[0]
	o := pts[idx[0]]
	rest := idx[1:]
	sort.SliceStable(rest, func(a, b int) bool {
		c := Cross(o, pts[rest[a]], pts[rest[b]])
		if c != 0 {
			return c > 0
		}
		return dist2(o, pts[rest[a]]) < dist2(o, pts[rest[b]])
	})
	stack := []int{idx[0]}
	for _, p := range rest {
		x.visit(p)
		for len(stack) >= 2 {
			a, b := stack[len(stack)-2], stack[len(stack)-1]
			x.compare(a, b, p)
			if Cross(pts[a], pts[b], pts[p]) > 0 {
				break
			}
			stack = stack[:len(stack)-1]
		}
		stack = append(stack, p)
		x.mark("hull", stack...)
	}
	return stack
}

// ConvexHullJarvis returns the convex hull with the Jarvis march (gift
// wrapping, 1973): from the lowest point, repeatedly pick the point such that
// every other point is to the left of the line to it. Output as for
// ConvexHullGraham.
//
// Time: O(n * h) for h hull points. Space: O(h).
func ConvexHullJarvis(pts []Point) []int { return jarvis(pts, nop) }

// TraceConvexHullJarvis is ConvexHullJarvis, reporting each candidate tested
// (Compare) and the hull as it is wrapped (Mark "hull").
func TraceConvexHullJarvis(pts []Point, r trace.Recorder) []int { return jarvis(pts, rec{r}) }

func jarvis(pts []Point, x rec) []int {
	idx := distinct(pts)
	if len(idx) < 3 {
		return idx
	}
	start := idx[0]
	for _, p := range idx {
		if pts[p].Y < pts[start].Y || (pts[p].Y == pts[start].Y && pts[p].X < pts[start].X) {
			start = p
		}
	}
	hull := []int{start}
	cur := start
	for {
		next := -1
		for _, p := range idx {
			if p == cur {
				continue
			}
			if next < 0 {
				next = p
				continue
			}
			x.compare(cur, next, p)
			c := Cross(pts[cur], pts[next], pts[p])
			// p is more clockwise than next, or as far round and further away.
			if c < 0 || (c == 0 && dist2(pts[cur], pts[p]) > dist2(pts[cur], pts[next])) {
				next = p
			}
		}
		if next == start {
			break
		}
		hull = append(hull, next)
		x.mark("hull", hull...)
		cur = next
		if len(hull) > len(idx) { // cannot happen with exact arithmetic; a guard all the same
			break
		}
	}
	return hull
}

// ConvexHullMonotone returns the convex hull with Andrew's monotone chain
// (1979): sort the points by x, then build the lower and upper hulls with the
// same left-turn stack as Graham's scan. Output as for ConvexHullGraham.
//
// Time: O(n log n). Space: O(n).
func ConvexHullMonotone(pts []Point) []int { return monotone(pts, nop) }

// TraceConvexHullMonotone is ConvexHullMonotone, reporting each turn test and
// the hull as each chain grows (Mark "hull").
func TraceConvexHullMonotone(pts []Point, r trace.Recorder) []int { return monotone(pts, rec{r}) }

func monotone(pts []Point, x rec) []int {
	idx := distinct(pts)
	if len(idx) < 3 {
		return idx
	}
	sort.Slice(idx, func(a, b int) bool {
		p, q := pts[idx[a]], pts[idx[b]]
		return p.X < q.X || (p.X == q.X && p.Y < q.Y)
	})
	var hull []int
	build := func(order []int, floor int) {
		for _, p := range order {
			x.visit(p)
			for len(hull) >= floor+2 {
				a, b := hull[len(hull)-2], hull[len(hull)-1]
				x.compare(a, b, p)
				if Cross(pts[a], pts[b], pts[p]) > 0 {
					break
				}
				hull = hull[:len(hull)-1]
			}
			hull = append(hull, p)
			x.mark("hull", hull...)
		}
	}
	build(idx, 0) // lower hull, left to right
	lower := len(hull) - 1
	rev := make([]int, len(idx)-1)
	for i := range rev {
		rev[i] = idx[len(idx)-2-i]
	}
	build(rev, lower)         // upper hull, right to left
	hull = hull[:len(hull)-1] // the last point repeats the first
	return rotateToLowest(pts, hull)
}

// rotateToLowest starts the hull at its lowest (then leftmost) point, as the
// other two hulls do.
func rotateToLowest(pts []Point, hull []int) []int {
	low := 0
	for i, p := range hull {
		q := pts[hull[low]]
		if pts[p].Y < q.Y || (pts[p].Y == q.Y && pts[p].X < q.X) {
			low = i
		}
	}
	return append(append([]int(nil), hull[low:]...), hull[:low]...)
}

// distinct returns the indexes of pts with each position once (the first
// index of repeated points).
func distinct(pts []Point) []int {
	seen := map[Point]bool{}
	var out []int
	for i, p := range pts {
		if !seen[p] {
			seen[p] = true
			out = append(out, i)
		}
	}
	return out
}
