package graphs

import (
	"container/heap"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// UniformCost finds the cheapest path from start to target with uniform-cost
// search: Dijkstra's algorithm that stops as soon as target is settled, the
// form search courses teach. It returns the path and its cost, or nil and -1
// when target cannot be reached. Weights must not be negative (ErrNegative).
//
// Time: O((V + E) log V). Space: O(V).
func UniformCost(g *Graph, start, target int) ([]int, int, error) {
	return heuristicSearch(g, nil, start, target, nop)
}

// TraceUniformCost is UniformCost, reporting each node added to the frontier
// (Push, Values[0] its cost so far), each node settled (Visit), and the path
// (Mark "path").
func TraceUniformCost(g *Graph, start, target int, r trace.Recorder) ([]int, int, error) {
	return heuristicSearch(g, nil, start, target, rec{r})
}

// heuristicSearch is the best-first search UniformCost and AStarGraph share:
// a node's priority is its cost so far plus h of it (nil h is 0).
func heuristicSearch(g *Graph, h func(int) int, start, target int, x rec) ([]int, int, error) {
	for _, n := range []int{start, target} {
		if err := g.checkNode(n); err != nil {
			return nil, -1, err
		}
	}
	for _, e := range g.Edges {
		if e.Weight < 0 {
			return nil, -1, ErrNegative
		}
	}
	if h == nil {
		h = func(int) int { return 0 }
	}
	adj := g.adjacency()
	cost, prev := filled(g.Len(), Infinity), filled(g.Len(), -1)
	done := make([]bool, g.Len())
	cost[start] = 0
	q := &minQueue{{start, h(start), 0}}
	x.r.Record(trace.Event{Kind: trace.Push, Indices: []int{start}, Values: []int{0}})
	pushed := 1
	for q.Len() > 0 {
		u := heap.Pop(q).(item).node
		if done[u] {
			continue
		}
		done[u] = true
		x.visit(u)
		if u == target {
			path := PathTo(prev, start, target)
			x.mark("path", path)
			return path, cost[u], nil
		}
		for _, a := range adj[u] {
			if c := cost[u] + a.weight; !done[a.to] && c < cost[a.to] {
				cost[a.to], prev[a.to] = c, u
				heap.Push(q, item{a.to, c + h(a.to), pushed})
				pushed++
				x.r.Record(trace.Event{Kind: trace.Push, Indices: []int{a.to}, Values: []int{c}})
			}
		}
	}
	return nil, -1, nil
}
