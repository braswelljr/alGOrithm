package pathfinding

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Dijkstra finds the cheapest path from g.Start to g.End, counting the cost of
// entering each cell, with Dijkstra's algorithm (Edsger Dijkstra, 1956):
// always explore the cell with the lowest known cost from the start, so when
// the end is explored its cost is final. It returns the path and its cost, or
// nil and -1.
//
// Time: O(V log V) for V cells, with a binary heap.
// Space: O(V).
func Dijkstra(g *Grid) ([]int, int) { return best(g, nil, false, rec{trace.Nop{}}) }

// TraceDijkstra is Dijkstra, reporting to r as the package describes; Push and
// Visit carry the cell's cost from the start in Values.
func TraceDijkstra(g *Grid, r trace.Recorder) ([]int, int) { return best(g, nil, false, rec{r}) }

// best is the search Dijkstra, A* and greedy best-first share. A cell's
// priority is its cost from the start plus h of it (h nil means 0: Dijkstra),
// or h alone when greedy. Ties go to the lower h, which sends A* toward the end
// instead of fanning out across equal-cost cells.
func best(g *Grid, h func(cell int) int, greedy bool, x rec) ([]int, int) {
	if h == nil {
		h = func(int) int { return 0 }
	}
	const unknown = -1
	dist := make([]int, len(g.Cost))
	for i := range dist {
		dist[i] = unknown
	}
	prev := unreached(len(g.Cost))
	done := make([]bool, len(g.Cost))
	var f frontier
	dist[g.Start] = 0
	f.Push(g.Start, h(g.Start), h(g.Start))
	x.push(g.Start, 0)
	for f.Len() > 0 {
		cur, _ := f.Pop()
		if done[cur] {
			continue // an older, costlier entry for a cell already explored
		}
		done[cur] = true
		x.visit(cur, dist[cur])
		if cur == g.End {
			p, _ := x.path(g, prev)
			return p, dist[cur]
		}
		for _, n := range g.Neighbors(cur) {
			d := dist[cur] + g.Cost[n]
			if done[n] || (dist[n] != unknown && d >= dist[n]) {
				continue
			}
			dist[n], prev[n] = d, cur
			priority := d + h(n)
			if greedy {
				priority = h(n)
			}
			f.Push(n, priority, h(n))
			x.push(n, d)
		}
	}
	return nil, -1
}
