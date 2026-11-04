package geometry

import "github.com/braswelljr/alGOrithm/pkg/trace"

// PolygonArea returns twice the signed area of the polygon whose vertices are
// pts in order, with the shoelace formula: positive when the vertices run
// counterclockwise, negative when clockwise. Twice the area is always a whole
// number for integer points, so the result is exact; divide by 2 for the area.
//
// Time: O(n). Space: O(1).
func PolygonArea(pts []Point) int64 {
	var s int64
	for i := range pts {
		j := (i + 1) % len(pts)
		s += int64(pts[i].X)*int64(pts[j].Y) - int64(pts[j].X)*int64(pts[i].Y)
	}
	return s
}

// PointInPolygon reports whether p is inside the polygon poly (vertices in
// order, either direction) by ray casting: count how many edges a ray from p
// to the right crosses; an odd count is inside. Points on an edge count as
// inside.
//
// Time: O(n). Space: O(1).
func PointInPolygon(poly []Point, p Point) bool {
	inside := false
	for i := range poly {
		a, b := poly[i], poly[(i+1)%len(poly)]
		if Cross(a, b, p) == 0 && onSegment(Segment{a, b}, p) {
			return true
		}
		if (a.Y > p.Y) != (b.Y > p.Y) {
			// The x where the edge crosses the ray's height, compared exactly:
			// p.X < a.X + (p.Y-a.Y)*(b.X-a.X)/(b.Y-a.Y).
			num := int64(p.Y-a.Y) * int64(b.X-a.X)
			den := int64(b.Y - a.Y)
			lhs := int64(p.X-a.X) * den
			if (den > 0 && lhs < num) || (den < 0 && lhs > num) {
				inside = !inside
			}
		}
	}
	return inside
}

// PointsInPolygon returns which of points lie inside poly.
//
// Time: O(n * m) for n points and m vertices. Space: O(n).
func PointsInPolygon(poly, points []Point) []bool { return pointsIn(poly, points, 0, nop) }

// TracePointsInPolygon is PointsInPolygon with the polygon and the points in
// one list (the polygon first, then the points from index first), reporting
// each point tested (Visit) and each one found inside (Mark "inside").
func TracePointsInPolygon(poly, points []Point, first int, r trace.Recorder) []bool {
	return pointsIn(poly, points, first, rec{r})
}

func pointsIn(poly, points []Point, first int, x rec) []bool {
	out := make([]bool, len(points))
	for i, p := range points {
		x.visit(first + i)
		if out[i] = PointInPolygon(poly, p); out[i] {
			x.mark("inside", first+i)
		}
	}
	return out
}
