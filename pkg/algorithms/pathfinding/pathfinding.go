// Package pathfinding finds paths through a grid: breadth-first and
// depth-first search, Dijkstra's algorithm, A* and greedy best-first search,
// one file each, over the Grid type in grid.go (walls, cells costing 1 to 9 to
// enter, a start and an end), which also parses grids from text and generates
// mazes.
//
// Every search X returns the path from the start to the end (cell indexes,
// row-major) and its cost, or nil and -1 when the end cannot be reached; TraceX
// reports each cell added to the frontier (Push), each cell explored (Visit),
// and finally the path (Mark "path"):
//
//	g, _ := pathfinding.Generate(pathfinding.Maze, 21, 31, 1)
//	path, cost := pathfinding.AStar(g)
package pathfinding

import "github.com/braswelljr/alGOrithm/pkg/trace"

type rec struct{ r trace.Recorder }

func (x rec) push(cell int, values ...int) {
	x.r.Record(trace.Event{Kind: trace.Push, Indices: []int{cell}, Values: values})
}

func (x rec) visit(cell int, values ...int) {
	x.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{cell}, Values: values})
}

func (x rec) path(g *Grid, prev []int) ([]int, int) {
	p := PathTo(prev, g.Start, g.End)
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: p, Note: "path"})
	return p, g.PathCost(p)
}

func unreached(n int) []int {
	prev := make([]int, n)
	for i := range prev {
		prev[i] = -1
	}
	return prev
}
