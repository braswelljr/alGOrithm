package pathfinding

import (
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// valid reports whether path walks from the start to the end of g, one step
// at a time, through open cells only.
func valid(g *Grid, path []int) bool {
	if len(path) < 1 || path[0] != g.Start || path[len(path)-1] != g.End {
		return false
	}
	for i, c := range path {
		if g.Cost[c] == 0 {
			return false
		}
		if i > 0 {
			a, b := path[i-1], c
			dr, dc := a/g.Cols-b/g.Cols, a%g.Cols-b%g.Cols
			if dr*dr+dc*dc != 1 {
				return false
			}
		}
	}
	return true
}

// TestUntracedPathfinders runs A*, depth-first and greedy best-first without a
// recorder on every kind of generated grid: each gives the traced function's
// answer, a valid path, and A* gives Dijkstra's (optimal) cost.
func TestUntracedPathfinders(t *testing.T) {
	for _, kind := range Kinds {
		for seed := int64(1); seed <= 4; seed++ {
			g, err := Generate(kind, 13, 21, seed)
			if err != nil {
				t.Fatalf("Generate(%s, %d): %v", kind, seed, err)
			}
			for _, c := range []struct {
				name   string
				plain  func(*Grid) ([]int, int)
				traced func(*Grid, trace.Recorder) ([]int, int)
			}{
				{"a-star", AStar, TraceAStar},
				{"depth-first", DepthFirst, TraceDepthFirst},
				{"greedy", Greedy, TraceGreedy},
			} {
				path, cost := c.plain(g)
				var log trace.Log
				tpath, tcost := c.traced(g, &log)
				if !reflect.DeepEqual(path, tpath) || cost != tcost {
					t.Errorf("%s on %s %d: untraced (%v, %d), traced (%v, %d)", c.name, kind, seed, path, cost, tpath, tcost)
				}
				if path != nil && !valid(g, path) {
					t.Errorf("%s on %s %d: %v is not a path from S to E", c.name, kind, seed, path)
				}
			}
			_, best := Dijkstra(g)
			if _, cost := AStar(g); cost != best {
				t.Errorf("A* on %s %d costs %d, Dijkstra %d", kind, seed, cost, best)
			}
		}
	}
}
