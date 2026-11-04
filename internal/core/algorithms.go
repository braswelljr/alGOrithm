package core

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/pathfinding"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// pathfindingEntries is every entry of the pathfinding category.
func pathfindingEntries() []Entry {
	return []Entry{
		aStarEntry(),
		breadthFirstSearchEntry(),
		depthFirstSearchEntry(),
		dijkstraEntry(),
		greedyBestFirstEntry(),
	}
}

var pathfindingMaze, _ = pathfinding.Generate(pathfinding.Maze, 15, 31, 1)

var pathfindingEx = &Input{Text: pathfindingMaze.String()}

func pathfindingMk(id, name, file, fn, summary, time string, traced func(*pathfinding.Grid, trace.Recorder) ([]int, int)) Entry {
	e := entry(id, name, "algorithms/pathfinding/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Example = summary, time, "O(V)", GridInput, true, pathfindingEx
	e.Limits = fmt.Sprintf("at most %d cells", pathfinding.MaxCells)
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		g, err := pathfinding.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		path, cost := traced(g, r)
		return Output{Grid: &GridOutput{Rows: g.Rows, Cols: g.Cols, Cost: g.Cost, Start: g.Start, End: g.End}, Path: path, Number: &cost}, nil
	}
	return e
}

func breadthFirstSearchEntry() Entry {
	return pathfindingMk("breadth-first-search", "Breadth-First Search", "breadthfirst.go", "BreadthFirst", "Explore every cell one step away, then two, and so on; the first path to the end has the fewest steps.", "O(V)", pathfinding.TraceBreadthFirst)
}

func depthFirstSearchEntry() Entry {
	return pathfindingMk("depth-first-search", "Depth-First Search", "depthfirst.go", "DepthFirst", "Follow one direction as far as it goes and back up at dead ends; finds a path, not the shortest.", "O(V)", pathfinding.TraceDepthFirst)
}

func dijkstraEntry() Entry {
	return pathfindingMk("dijkstra", "Dijkstra's Algorithm", "dijkstra.go", "Dijkstra", "Always explore the cell cheapest to reach from the start; finds the cheapest path (Dijkstra, 1956).", "O(V log V)", pathfinding.TraceDijkstra)
}

func aStarEntry() Entry {
	return pathfindingMk("a-star", "A* Search", "astar.go", "AStar", "Dijkstra's algorithm guided by an estimate of the cost still to go; finds the cheapest path, exploring less.", "O(V log V) worst", pathfinding.TraceAStar)
}

func greedyBestFirstEntry() Entry {
	return pathfindingMk("greedy-best-first", "Greedy Best-First Search", "greedy.go", "Greedy", "Always explore the cell that looks closest to the end; fast, but not always cheapest.", "O(V log V)", pathfinding.TraceGreedy)
}

const (
	intervalsExample = "name start end\na 1 4\nb 3 5\nc 0 6\nd 5 7\ne 3 9\nf 5 9\ng 6 10\nh 8 11\ni 8 12\nj 2 14\nk 12 16"
	itemsExample     = "name weight value\ngold 10 60\nsilver 20 100\nbronze 30 120"
	jobsExample      = "name deadline profit\na 2 100\nb 1 19\nc 2 27\nd 1 25\ne 3 15"
	sudokuExample    = "53..7....\n6..195...\n.98....6.\n8...6...3\n4..8.3..1\n7...2...6\n.6....28.\n...419..5\n....8..79"
)

// parseCall writes "name, err := pkg.Parse(`text`)" setup for snippets.
func parseCall(v, fn, text string) string {
	return fmt.Sprintf("%s, err := %s(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", v, fn, "`"+text+"`")
}

func chars(b []byte) []string {
	out := make([]string, len(b))
	for i, c := range b {
		out[i] = string(rune(c))
	}
	return out
}
