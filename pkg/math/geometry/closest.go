package geometry

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ClosestPair returns the indexes of the two points of pts closest to each
// other, and their squared distance, by divide and conquer (Shamos and Hoey,
// 1975): split by x, solve each half, then check only the points within the
// best distance of the dividing line, sorted by y, each against a few
// neighbors. With fewer than two points it returns -1, -1, 0.
//
// Time: O(n log^2 n) as written (a sort per strip). Space: O(n).
func ClosestPair(pts []Point) (i, j int, dist2 int64) { return closest(pts, nop) }

// TraceClosestPair is ClosestPair, reporting each pair measured (Compare) and
// each new best pair (Mark "pair").
func TraceClosestPair(pts []Point, r trace.Recorder) (int, int, int64) { return closest(pts, rec{r}) }

func closest(pts []Point, x rec) (int, int, int64) {
	if len(pts) < 2 {
		return -1, -1, 0
	}
	idx := make([]int, len(pts))
	for k := range idx {
		idx[k] = k
	}
	sort.Slice(idx, func(a, b int) bool {
		p, q := pts[idx[a]], pts[idx[b]]
		return p.X < q.X || (p.X == q.X && p.Y < q.Y)
	})
	bi, bj, best := -1, -1, int64(-1)
	try := func(a, b int) {
		x.compare(a, b)
		if d := dist2(pts[a], pts[b]); best < 0 || d < best {
			bi, bj, best = a, b, d
			x.mark("pair", a, b)
		}
	}
	var solve func(lo, hi int)
	solve = func(lo, hi int) {
		if hi-lo <= 3 {
			for a := lo; a < hi; a++ {
				for b := a + 1; b < hi; b++ {
					try(idx[a], idx[b])
				}
			}
			return
		}
		mid := (lo + hi) / 2
		midX := pts[idx[mid]].X
		solve(lo, mid)
		solve(mid, hi)
		var strip []int
		for k := lo; k < hi; k++ {
			dx := int64(pts[idx[k]].X - midX)
			if dx*dx < best {
				strip = append(strip, idx[k])
			}
		}
		sort.Slice(strip, func(a, b int) bool { return pts[strip[a]].Y < pts[strip[b]].Y })
		for a := range strip {
			for b := a + 1; b < len(strip); b++ {
				dy := int64(pts[strip[b]].Y - pts[strip[a]].Y)
				if dy*dy >= best {
					break // the rest are further away in y alone
				}
				try(strip[a], strip[b])
			}
		}
	}
	solve(0, len(idx))
	if bi > bj {
		bi, bj = bj, bi
	}
	return bi, bj, best
}
