package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// AStarGraph finds the cheapest path from start to target with A* search
// (Hart, Nilsson and Raphael, 1968): uniform-cost search guided by an
// estimate of the cost still to go, here the straight-line distance between
// node positions (ParsePositions), so it explores toward the target. The path
// is still the cheapest when every edge weighs at least the straight-line
// length between its ends; nodes without a position estimate 0, and with no
// positions at all it is uniform-cost search. It returns the path and its
// cost, or nil and -1.
//
// Time: O((V + E) log V) in the worst case, usually far less. Space: O(V).
func AStarGraph(g *Graph, pos Positions, start, target int) ([]int, int, error) {
	return heuristicSearch(g, func(v int) int { return pos.estimate(v, target) }, start, target, nop)
}

// TraceAStarGraph is AStarGraph, reporting what TraceUniformCost does.
func TraceAStarGraph(g *Graph, pos Positions, start, target int, r trace.Recorder) ([]int, int, error) {
	return heuristicSearch(g, func(v int) int { return pos.estimate(v, target) }, start, target, rec{r})
}
