package pathfinding

import "github.com/braswelljr/alGOrithm/pkg/trace"

// AStar finds the cheapest path from g.Start to g.End with A* search (Hart,
// Nilsson and Raphael, 1968): Dijkstra's algorithm, but each cell's priority
// adds an estimate of the cost still to go, so the search heads for the end.
// The estimate is the Manhattan distance, which never overestimates (every
// step costs at least 1), so the path is still the cheapest. It returns the
// path and its cost, or nil and -1.
//
// Time: O(V log V) in the worst case for V cells; usually far fewer cells are
// explored than by Dijkstra's algorithm.
// Space: O(V).
func AStar(g *Grid) ([]int, int) { return TraceAStar(g, trace.Nop{}) }

// TraceAStar is AStar, reporting what TraceDijkstra does.
func TraceAStar(g *Grid, r trace.Recorder) ([]int, int) {
	return best(g, func(cell int) int { return g.Distance(cell, g.End) }, false, rec{r})
}
