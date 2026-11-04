package spatial_test

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/spatial"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func sorted(ps []spatial.Point) []spatial.Point {
	out := append([]spatial.Point(nil), ps...)
	sort.Slice(out, func(i, j int) bool { return out[i].X < out[j].X || (out[i].X == out[j].X && out[i].Y < out[j].Y) })
	return out
}

func same(a, b []spatial.Point) bool {
	a, b = sorted(a), sorted(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestKDTree(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	var pts []spatial.Point
	for i := 0; i < 300; i++ {
		pts = append(pts, spatial.Point{X: r.Intn(200) - 100, Y: r.Intn(200) - 100})
	}
	built := spatial.NewKDTree(pts[:150]...)
	for _, p := range pts[150:] {
		built.Insert(p)
	}
	uniq := map[spatial.Point]bool{}
	for _, p := range pts {
		uniq[p] = true
	}
	if built.Len() != len(uniq) {
		t.Fatalf("Len %d, want %d", built.Len(), len(uniq))
	}
	for i := 0; i < 500; i++ {
		q := spatial.Point{X: r.Intn(240) - 120, Y: r.Intn(240) - 120}
		best := -1
		for p := range uniq {
			if d := spatial.Dist2(p, q); best < 0 || d < best {
				best = d
			}
		}
		if _, d, _, ok := built.Nearest(q); !ok || d != best {
			t.Fatalf("Nearest(%v) at distance² %d, want %d", q, d, best)
		}
		x, y := r.Intn(200)-100, r.Intn(200)-100
		rect := spatial.Rect{Min: spatial.Point{X: x, Y: y}, Max: spatial.Point{X: x + r.Intn(60), Y: y + r.Intn(60)}}
		var want []spatial.Point
		for p := range uniq {
			if rect.Contains(p) {
				want = append(want, p)
			}
		}
		if got := built.InRect(rect); !same(got, want) {
			t.Fatalf("InRect(%v) = %v, want %v", rect, got, want)
		}
	}
	if _, _, _, ok := spatial.NewKDTree().Nearest(spatial.Point{}); ok {
		t.Error("an empty tree found a point")
	}
}

func TestQuadTree(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	bounds := spatial.Rect{Min: spatial.Point{X: -64, Y: -64}, Max: spatial.Point{X: 63, Y: 63}}
	q := spatial.NewQuadTree(bounds, 3)
	uniq := map[spatial.Point]bool{}
	for i := 0; i < 400; i++ {
		p := spatial.Point{X: r.Intn(140) - 70, Y: r.Intn(140) - 70}
		want := bounds.Contains(p) && !uniq[p]
		if q.Insert(p) != want {
			t.Fatalf("Insert(%v) disagreed", p)
		}
		if want {
			uniq[p] = true
		}
	}
	for i := 0; i < 20; i++ { // a pile of identical points must not split forever
		q.Insert(spatial.Point{X: 1, Y: 1})
	}
	uniq[spatial.Point{X: 1, Y: 1}] = true
	if q.Len() != len(uniq) {
		t.Fatalf("Len %d, want %d", q.Len(), len(uniq))
	}
	for i := 0; i < 300; i++ {
		x, y := r.Intn(128)-64, r.Intn(128)-64
		rect := spatial.Rect{Min: spatial.Point{X: x, Y: y}, Max: spatial.Point{X: x + r.Intn(40), Y: y + r.Intn(40)}}
		var want []spatial.Point
		for p := range uniq {
			if rect.Contains(p) {
				want = append(want, p)
			}
		}
		if got, _ := q.Query(rect); !same(got, want) {
			t.Fatalf("Query(%v) = %v, want %v", rect, got, want)
		}
	}
}

func TestKDTreeSnapshot(t *testing.T) {
	// the classic example: the median split puts (7,2) at the root
	kd := spatial.NewKDTree(spatial.Point{X: 2, Y: 3}, spatial.Point{X: 5, Y: 4}, spatial.Point{X: 9, Y: 6},
		spatial.Point{X: 4, Y: 7}, spatial.Point{X: 8, Y: 1}, spatial.Point{X: 7, Y: 2}, spatial.Point{X: 2, Y: 3})
	if kd.Len() != 6 {
		t.Errorf("len %d, the duplicate is kept once", kd.Len())
	}
	want := trace.Snapshot{
		Layout: "tree",
		Nodes: []trace.Node{
			{ID: 0, Label: "(7,2) x", Role: "root", Parent: -1},
			{ID: 1, Label: "(5,4) y", Parent: 0},
			{ID: 2, Label: "(2,3) x", Parent: 1},
			{ID: 3, Label: "(4,7) x", Parent: 1},
			{ID: 4, Label: "(9,6) y", Parent: 0},
			{ID: 5, Label: "(8,1) x", Parent: 4},
		},
		Edges: []trace.Edge{
			{From: 0, To: 1, Label: "left"}, {From: 1, To: 2, Label: "left"}, {From: 1, To: 3, Label: "right"},
			{From: 0, To: 4, Label: "right"}, {From: 4, To: 5, Label: "left"},
		},
	}
	if got := kd.Snapshot(); !reflect.DeepEqual(got, want) {
		t.Errorf("snapshot %+v\nwant %+v", got, want)
	}
	if e := spatial.NewKDTree().Snapshot(); e.Nodes == nil || len(e.Nodes) != 0 || len(e.Edges) != 0 {
		t.Errorf("an empty tree's snapshot %+v", e)
	}
}

func TestQuadTreeSnapshot(t *testing.T) {
	q := spatial.NewQuadTree(spatial.Rect{Min: spatial.Point{}, Max: spatial.Point{X: 7, Y: 7}}, 0) // capacity becomes 1
	if !q.Insert(spatial.Point{X: 1, Y: 1}) || !q.Insert(spatial.Point{X: 6, Y: 6}) {
		t.Fatal("inserts inside the bounds failed")
	}
	if q.Insert(spatial.Point{X: 6, Y: 6}) || q.Insert(spatial.Point{X: 8, Y: 0}) {
		t.Error("a duplicate or an outside point was inserted")
	}
	if q.Contains(spatial.Point{X: 9, Y: 9}) || q.Contains(spatial.Point{X: 2, Y: 2}) || !q.Contains(spatial.Point{X: 6, Y: 6}) {
		t.Error("Contains")
	}
	want := trace.Snapshot{
		Layout: "tree",
		Nodes: []trace.Node{
			{ID: 0, Label: "(0,0)-(7,7)", Role: "root", Parent: -1},
			{ID: 1, Label: "(0,0)-(3,3) [(1,1)]", Parent: 0},
			{ID: 2, Label: "(4,0)-(7,3) []", Parent: 0},
			{ID: 3, Label: "(0,4)-(3,7) []", Parent: 0},
			{ID: 4, Label: "(4,4)-(7,7) [(6,6)]", Parent: 0},
		},
		Edges: []trace.Edge{{From: 0, To: 1, Label: "SW"}, {From: 0, To: 2, Label: "SE"}, {From: 0, To: 3, Label: "NW"}, {From: 0, To: 4, Label: "NE"}},
	}
	if got := q.Snapshot(); !reflect.DeepEqual(got, want) {
		t.Errorf("snapshot %+v\nwant %+v", got, want)
	}
	if q.Len() != 2 {
		t.Errorf("len %d", q.Len())
	}
	one := spatial.NewQuadTree(spatial.Rect{}, 1)
	one.Insert(spatial.Point{})
	if s := one.Snapshot(); len(s.Nodes) != 1 || s.Nodes[0].Label != "(0,0)-(0,0) [(0,0)]" {
		t.Errorf("a one-cell tree %+v", s)
	}
}

func TestPointAndRectStrings(t *testing.T) {
	p := spatial.Point{X: -1, Y: 2}
	if p.String() != "(-1,2)" {
		t.Errorf("point %q", p.String())
	}
	r := spatial.Rect{Min: p, Max: spatial.Point{X: 3, Y: 4}}
	if r.String() != "(-1,2)-(3,4)" {
		t.Errorf("rect %q", r.String())
	}
	if !r.Contains(p) || r.Contains(spatial.Point{X: 4, Y: 4}) || spatial.Dist2(p, spatial.Point{X: 2, Y: 6}) != 25 {
		t.Error("Contains or Dist2")
	}
}
