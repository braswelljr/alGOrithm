package graphs

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// BeamSearch looks for a path from start to target keeping only the width
// most promising nodes of each level, ranked by the straight-line estimate to
// target (ParsePositions; nodes without positions rank by cost so far). It
// uses far less memory than breadth-first search but can miss the target or
// return a costly path: it is a heuristic, not a guarantee. It returns the
// path and its cost, or nil and -1 when the beam lost the target. width below
// 1 is 1.
//
// Time: O(d * w * b log(w * b)) for depth d, width w, branching b. Space:
// O(V) for the parents.
func BeamSearch(g *Graph, pos Positions, start, target, width int) ([]int, int, error) {
	return beam(g, pos, start, target, width, nop)
}

// TraceBeamSearch is BeamSearch, reporting each node kept in the beam (Push),
// each node expanded (Visit), and the path (Mark "path").
func TraceBeamSearch(g *Graph, pos Positions, start, target, width int, r trace.Recorder) ([]int, int, error) {
	return beam(g, pos, start, target, width, rec{r})
}

func beam(g *Graph, pos Positions, start, target, width int, x rec) ([]int, int, error) {
	for _, n := range []int{start, target} {
		if err := g.checkNode(n); err != nil {
			return nil, -1, err
		}
	}
	width = max(width, 1)
	adj := g.adjacency()
	prev, cost := filled(g.Len(), -1), filled(g.Len(), Infinity)
	seen := make([]bool, g.Len())
	seen[start], cost[start] = true, 0
	level := []int{start}
	x.push(start)
	for len(level) > 0 {
		var next []int
		for _, u := range level {
			x.visit(u)
			if u == target {
				path := PathTo(prev, start, target)
				x.mark("path", path)
				return path, cost[target], nil
			}
			for _, a := range adj[u] {
				if !seen[a.to] {
					seen[a.to], prev[a.to], cost[a.to] = true, u, cost[u]+a.weight
					next = append(next, a.to)
				}
			}
		}
		sort.SliceStable(next, func(i, j int) bool {
			hi, hj := pos.estimate(next[i], target), pos.estimate(next[j], target)
			if hi != hj {
				return hi < hj
			}
			return cost[next[i]] < cost[next[j]]
		})
		if len(next) > width {
			next = next[:width]
		}
		for _, v := range next {
			x.push(v)
		}
		level = next
	}
	return nil, -1, nil
}
