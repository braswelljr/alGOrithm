package pathfinding_test

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/pathfinding"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type finder struct {
	name     string
	trace    func(*pathfinding.Grid, trace.Recorder) ([]int, int)
	cheapest bool // the path's cost is the least possible
	fewest   bool // the path has the fewest steps
}

var finders = []finder{
	{"bfs", pathfinding.TraceBreadthFirst, false, true},
	{"dfs", pathfinding.TraceDepthFirst, false, false},
	{"dijkstra", pathfinding.TraceDijkstra, true, false},
	{"astar", pathfinding.TraceAStar, true, false},
	{"greedy", pathfinding.TraceGreedy, false, false},
}

// valid checks that path walks from start to end through open neighbors.
func valid(t *testing.T, g *pathfinding.Grid, path []int) {
	t.Helper()
	if path[0] != g.Start || path[len(path)-1] != g.End {
		t.Fatalf("path %v does not run start %d to end %d", path, g.Start, g.End)
	}
	for i := 1; i < len(path); i++ {
		if g.Cost[path[i]] == 0 || g.Distance(path[i-1], path[i]) != 1 {
			t.Fatalf("step %d -> %d is not a move to an open neighbor", path[i-1], path[i])
		}
	}
}

func TestPathfinders(t *testing.T) {
	for _, kind := range pathfinding.Kinds {
		for seed := int64(1); seed <= 4; seed++ {
			g, err := pathfinding.Generate(kind, 15, 21, seed)
			if err != nil {
				t.Fatal(err)
			}
			_, best := pathfinding.Dijkstra(g)
			steps, _ := pathfinding.BreadthFirst(g)
			for _, f := range finders {
				var log trace.Log
				path, cost := f.trace(g, &log)
				if path == nil {
					if best != -1 {
						t.Errorf("%s/%s/%d: no path, but one exists", f.name, kind, seed)
					}
					continue
				}
				valid(t, g, path)
				if cost != g.PathCost(path) {
					t.Errorf("%s: reported cost %d, path costs %d", f.name, cost, g.PathCost(path))
				}
				if f.cheapest && cost != best {
					t.Errorf("%s/%s/%d: cost %d, cheapest is %d", f.name, kind, seed, cost, best)
				}
				if f.fewest && len(path) != len(steps) {
					t.Errorf("%s: %d steps, fewest is %d", f.name, len(path), len(steps))
				}
				if log.Count(trace.Visit) == 0 || log.Events[len(log.Events)-1].Note != "path" {
					t.Errorf("%s: events do not end with the path", f.name)
				}
			}
		}
	}
}

func TestNoPath(t *testing.T) {
	g, err := pathfinding.Parse("S#.\n.#.\n.#E")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range finders {
		if path, cost := f.trace(g, trace.Nop{}); path != nil || cost != -1 {
			t.Errorf("%s found %v, %d through a wall", f.name, path, cost)
		}
	}
}

func TestAStarExploresLessThanDijkstra(t *testing.T) {
	g, _ := pathfinding.Generate(pathfinding.Open, 30, 30, 1)
	var d, a trace.Log
	pathfinding.TraceDijkstra(g, &d)
	pathfinding.TraceAStar(g, &a)
	if a.Count(trace.Visit) >= d.Count(trace.Visit) {
		t.Errorf("A* explored %d cells, Dijkstra %d", a.Count(trace.Visit), d.Count(trace.Visit))
	}
}
