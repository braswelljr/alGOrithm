package pathfinding

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Greedy finds a path from g.Start to g.End with greedy best-first search:
// always explore the cell that looks closest to the end, ignoring the cost so
// far. It is fast and usually direct, but the path need not be the cheapest.
// It returns the path and its cost, or nil and -1.
//
// Time: O(V log V) in the worst case for V cells.
// Space: O(V).
func Greedy(g *Grid) ([]int, int) { return TraceGreedy(g, trace.Nop{}) }

// TraceGreedy is Greedy, reporting what TraceDijkstra does.
func TraceGreedy(g *Grid, r trace.Recorder) ([]int, int) {
	return best(g, func(cell int) int { return g.Distance(cell, g.End) }, true, rec{r})
}
