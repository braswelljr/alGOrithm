package geometry_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/math/geometry"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func randomPoints(r *rand.Rand, n, span int) []geometry.Point {
	pts := make([]geometry.Point, n)
	for i := range pts {
		pts[i] = geometry.Point{X: r.Intn(span), Y: r.Intn(span)}
	}
	return pts
}

// bruteHull returns the hull's corners the slow way: a point is not a corner
// when it lies on a segment between two other points or inside a triangle of
// three others.
func bruteHull(pts []geometry.Point) map[geometry.Point]bool {
	uniq := map[geometry.Point]bool{}
	for _, p := range pts {
		uniq[p] = true
	}
	var all []geometry.Point
	for p := range uniq {
		all = append(all, p)
	}
	out := map[geometry.Point]bool{}
	for _, p := range all {
		corner := true
		for i := 0; i < len(all) && corner; i++ {
			for j := i + 1; j < len(all) && corner; j++ {
				a, b := all[i], all[j]
				if a == p || b == p {
					continue
				}
				if geometry.Cross(a, b, p) == 0 && geometry.Intersect(geometry.Segment{A: a, B: b}, geometry.Segment{A: p, B: p}) {
					corner = false
				}
				for k := j + 1; k < len(all) && corner; k++ {
					c := all[k]
					if c == p {
						continue
					}
					if geometry.PointInPolygon([]geometry.Point{a, b, c}, p) && geometry.Cross(a, b, c) != 0 {
						corner = false
					}
				}
			}
		}
		if corner {
			out[p] = true
		}
	}
	if len(all) < 3 {
		for _, p := range all {
			out[p] = true
		}
	}
	return out
}

func TestHullsAgreeWithBruteForce(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	hulls := map[string]func([]geometry.Point) []int{
		"graham": geometry.ConvexHullGraham, "jarvis": geometry.ConvexHullJarvis, "monotone": geometry.ConvexHullMonotone,
	}
	for round := 0; round < 60; round++ {
		pts := randomPoints(r, 3+r.Intn(25), 12) // a small span: many duplicates and collinear points
		want := bruteHull(pts)
		var first []int
		for name, f := range hulls {
			got := f(pts)
			set := map[geometry.Point]bool{}
			for _, i := range got {
				set[pts[i]] = true
			}
			if !reflect.DeepEqual(set, want) {
				t.Fatalf("round %d %s: hull %v, want corners %v (points %v)", round, name, got, want, pts)
			}
			if len(got) >= 3 && geometry.PolygonArea(hullPoints(pts, got)) <= 0 {
				t.Fatalf("%s: hull is not counterclockwise", name)
			}
			if first == nil {
				first = got
			} else if !reflect.DeepEqual(hullPoints(pts, got), hullPoints(pts, first)) {
				t.Fatalf("round %d: hulls differ in order: %v vs %v", round, hullPoints(pts, got), hullPoints(pts, first))
			}
		}
	}
}

func hullPoints(pts []geometry.Point, idx []int) []geometry.Point {
	out := make([]geometry.Point, len(idx))
	for i, k := range idx {
		out[i] = pts[k]
	}
	return out
}

func TestHullTraceGrows(t *testing.T) {
	pts, _ := geometry.ParsePoints("0 0\n4 0\n2 1\n4 4\n0 4\n2 2")
	var log trace.Log
	hull := geometry.TraceConvexHullGraham(pts, &log)
	if len(hull) != 4 || log.Count(trace.Mark) == 0 {
		t.Fatalf("hull %v, %d marks", hull, log.Count(trace.Mark))
	}
	last := log.Events[len(log.Events)-1]
	if last.Note != "hull" || !reflect.DeepEqual(last.Indices, hull) {
		t.Errorf("last event %+v, want the final hull %v", last, hull)
	}
}

func TestClosestPair(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for round := 0; round < 100; round++ {
		pts := randomPoints(r, 2+r.Intn(60), 1000)
		var best int64 = -1
		for i := range pts {
			for j := i + 1; j < len(pts); j++ {
				dx, dy := int64(pts[i].X-pts[j].X), int64(pts[i].Y-pts[j].Y)
				if d := dx*dx + dy*dy; best < 0 || d < best {
					best = d
				}
			}
		}
		i, j, d := geometry.ClosestPair(pts)
		dx, dy := int64(pts[i].X-pts[j].X), int64(pts[i].Y-pts[j].Y)
		if d != best || dx*dx+dy*dy != d || i >= j {
			t.Fatalf("round %d: pair %d %d at %d, brute force %d", round, i, j, d, best)
		}
	}
	if i, j, _ := geometry.ClosestPair([]geometry.Point{{1, 1}}); i != -1 || j != -1 {
		t.Error("one point gave a pair")
	}
}

func TestSegments(t *testing.T) {
	pts, _ := geometry.ParsePoints("0 0\n4 4\n0 4\n4 0\n5 5\n6 6\n2 2\n8 8\n10 0\n10 3")
	segs, err := geometry.Segments(pts)
	if err != nil {
		t.Fatal(err)
	}
	got := geometry.SegmentsIntersect(segs)
	// 0 crosses 1 at (2,2); 0 and 3 overlap on (2,2)-(4,4); 2 lies inside 3; 4 touches nothing.
	want := [][2]int{{0, 1}, {0, 3}, {1, 3}, {2, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("intersections %v, want %v", got, want)
	}
	if _, err := geometry.Segments(pts[:3]); err == nil {
		t.Error("an odd number of points made segments")
	}
}

func TestPolygon(t *testing.T) {
	square := []geometry.Point{{0, 0}, {4, 0}, {4, 4}, {0, 4}}
	if a := geometry.PolygonArea(square); a != 32 {
		t.Errorf("twice the square's area = %d", a)
	}
	if geometry.PolygonArea([]geometry.Point{{0, 4}, {4, 4}, {4, 0}, {0, 0}}) != -32 {
		t.Error("a clockwise polygon should have a negative area")
	}
	got := geometry.PointsInPolygon(square, []geometry.Point{{2, 2}, {5, 2}, {4, 2}, {0, 0}, {-1, -1}})
	if !reflect.DeepEqual(got, []bool{true, false, true, true, false}) {
		t.Errorf("inside = %v", got)
	}
	concave := []geometry.Point{{0, 0}, {6, 0}, {6, 6}, {3, 2}, {0, 6}}
	if geometry.PointInPolygon(concave, geometry.Point{X: 3, Y: 4}) || !geometry.PointInPolygon(concave, geometry.Point{X: 1, Y: 1}) {
		t.Error("concave polygon test wrong")
	}
}

func TestParsePoints(t *testing.T) {
	for _, bad := range []string{"", "1", "1 2 3", "a b", "1 99999999999"} {
		if _, err := geometry.ParsePoints(bad); err == nil {
			t.Errorf("ParsePoints(%q) succeeded", bad)
		}
	}
	if pts, err := geometry.ParsePoints("# comment\n1, 2\n\n3\t4"); err != nil || len(pts) != 2 {
		t.Errorf("ParsePoints = %v, %v", pts, err)
	}
}
