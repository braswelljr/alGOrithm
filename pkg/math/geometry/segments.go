package geometry

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Segment is a line segment between two points.
type Segment struct{ A, B Point }

// Segments pairs consecutive points into segments: points 0-1, 2-3, ... An
// odd number of points is an error.
func Segments(pts []Point) ([]Segment, error) {
	if len(pts)%2 != 0 {
		return nil, fmt.Errorf("%w: segments need an even number of points (two per segment)", ErrInput)
	}
	out := make([]Segment, len(pts)/2)
	for i := range out {
		out[i] = Segment{pts[2*i], pts[2*i+1]}
	}
	return out, nil
}

// Intersect reports whether segments s and t share at least one point,
// touching ends and collinear overlaps included, with orientation tests.
func Intersect(s, t Segment) bool {
	d1, d2 := sign(Cross(t.A, t.B, s.A)), sign(Cross(t.A, t.B, s.B))
	d3, d4 := sign(Cross(s.A, s.B, t.A)), sign(Cross(s.A, s.B, t.B))
	if d1*d2 < 0 && d3*d4 < 0 {
		return true // a proper crossing
	}
	return (d1 == 0 && onSegment(t, s.A)) || (d2 == 0 && onSegment(t, s.B)) ||
		(d3 == 0 && onSegment(s, t.A)) || (d4 == 0 && onSegment(s, t.B))
}

// onSegment reports whether p, known to be collinear with s, lies on it.
func onSegment(s Segment, p Point) bool {
	return min(s.A.X, s.B.X) <= p.X && p.X <= max(s.A.X, s.B.X) &&
		min(s.A.Y, s.B.Y) <= p.Y && p.Y <= max(s.A.Y, s.B.Y)
}

func sign(v int64) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

// SegmentsIntersect returns every pair of segments (indexes into segs, i < j)
// that intersect, testing all pairs. A sweep line (Bentley-Ottmann) is faster
// when there are few intersections; this one is the one to read first.
//
// Time: O(n^2). Space: O(k) for k intersecting pairs.
func SegmentsIntersect(segs []Segment) [][2]int { return segmentsIntersect(segs, nop) }

// TraceSegmentsIntersect is SegmentsIntersect, reporting each pair tested
// (Compare, by segment index) and each intersecting pair (Mark "pair").
func TraceSegmentsIntersect(segs []Segment, r trace.Recorder) [][2]int {
	return segmentsIntersect(segs, rec{r})
}

func segmentsIntersect(segs []Segment, x rec) [][2]int {
	var out [][2]int
	for i := range segs {
		for j := i + 1; j < len(segs); j++ {
			x.compare(i, j)
			if Intersect(segs[i], segs[j]) {
				out = append(out, [2]int{i, j})
				x.mark("pair", i, j)
			}
		}
	}
	return out
}
