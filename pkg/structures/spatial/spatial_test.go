package spatial_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/spatial"
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
