package graphs

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ArticulationPoints returns the nodes whose removal disconnects their part of
// an undirected graph (cut vertices), in ascending order, with Tarjan's
// low-link depth-first search (1973): a node is one when a child's subtree
// cannot reach above it except through it.
//
// Time: O(V + E). Space: O(V).
func ArticulationPoints(g *Graph) ([]int, error) {
	points, _, err := lowLink(g, nop)
	return points, err
}

// TraceArticulationPoints is ArticulationPoints, reporting each node
// discovered (Visit) and each articulation point (Mark "group", Values[0] 1).
func TraceArticulationPoints(g *Graph, r trace.Recorder) ([]int, error) {
	points, _, err := lowLink(g, rec{r})
	return points, err
}

// Bridges returns the edges whose removal disconnects their part of an
// undirected graph, as edge indexes in ascending order, from the same
// low-link search: an edge to a child is one when the child's subtree cannot
// reach the parent or above any other way. One of two parallel edges is never
// a bridge.
//
// Time: O(V + E). Space: O(V).
func Bridges(g *Graph) ([]int, error) {
	_, bridges, err := lowLink(g, nop)
	return bridges, err
}

// TraceBridges is Bridges, reporting each node discovered (Visit) and each
// bridge (Mark "tree", Indices[0] the edge).
func TraceBridges(g *Graph, r trace.Recorder) ([]int, error) {
	_, bridges, err := lowLink(g, rec{r})
	return bridges, err
}

func lowLink(g *Graph, x rec) (points, bridges []int, err error) {
	if g.Directed {
		return nil, nil, ErrDirected
	}
	adj := g.adjacency()
	disc, low := filled(g.Len(), -1), make([]int, g.Len())
	isPoint := make([]bool, g.Len())
	clock := 0
	var visit func(u, viaEdge int)
	visit = func(u, viaEdge int) {
		disc[u], low[u] = clock, clock
		clock++
		x.visit(u)
		children := 0
		for _, a := range adj[u] {
			if a.edge == viaEdge {
				continue // the edge we came in by, not a parallel twin
			}
			if disc[a.to] >= 0 {
				low[u] = min(low[u], disc[a.to])
				continue
			}
			children++
			visit(a.to, a.edge)
			low[u] = min(low[u], low[a.to])
			if low[a.to] > disc[u] {
				bridges = append(bridges, a.edge)
				x.mark("tree", []int{a.edge})
			}
			if viaEdge >= 0 && low[a.to] >= disc[u] {
				isPoint[u] = true
			}
		}
		if viaEdge < 0 && children > 1 {
			isPoint[u] = true
		}
	}
	for u := range disc {
		if disc[u] < 0 {
			visit(u, -1)
		}
	}
	for u, p := range isPoint {
		if p {
			points = append(points, u)
			x.mark("group", []int{u}, 1)
		}
	}
	sort.Ints(bridges)
	return points, bridges, nil
}
