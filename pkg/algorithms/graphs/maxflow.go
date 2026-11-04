package graphs

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MaxFlow returns the maximum flow from source to sink, treating each edge's
// weight as its capacity, with the Edmonds-Karp algorithm (1972): while a path
// with spare capacity exists, push flow along the shortest one (found
// breadth-first). It also returns the flow on each edge: positive from From
// to To; on an undirected graph, negative when it runs To to From.
// Capacities must not be negative.
//
// Time: O(V * E^2). Space: O(V + E).
func MaxFlow(g *Graph, source, sink int) (total int, flow []int, err error) {
	return maxFlow(g, source, sink, nop)
}

// TraceMaxFlow is MaxFlow, reporting each augmenting path (Mark "path") and
// each edge's new flow (Mark "flow", Indices[0] the edge, Values[0] the flow).
func TraceMaxFlow(g *Graph, source, sink int, r trace.Recorder) (total int, flow []int, err error) {
	return maxFlow(g, source, sink, rec{r})
}

func maxFlow(g *Graph, source, sink int, x rec) (int, []int, error) {
	for _, n := range []int{source, sink} {
		if err := g.checkNode(n); err != nil {
			return 0, nil, err
		}
	}
	if source == sink {
		return 0, nil, fmt.Errorf("%w: the source is the sink", ErrInput)
	}
	// Residual arcs in pairs: 2i runs From to To, 2i+1 back.
	type arc struct{ to, cap int }
	arcs := make([]arc, 0, 2*len(g.Edges))
	adj := make([][]int, g.Len())
	for _, e := range g.Edges {
		if e.Weight < 0 {
			return 0, nil, fmt.Errorf("%w: capacities must not be negative", ErrInput)
		}
		back := 0
		if !g.Directed {
			back = e.Weight
		}
		adj[e.From] = append(adj[e.From], len(arcs))
		arcs = append(arcs, arc{e.To, e.Weight})
		adj[e.To] = append(adj[e.To], len(arcs))
		arcs = append(arcs, arc{e.From, back})
	}
	flowOf := func(i int) int { return g.Edges[i].Weight - arcs[2*i].cap }
	total := 0
	for {
		via := filled(g.Len(), -1) // the arc each node was reached by
		via[source] = -2
		queue := []int{source}
		for len(queue) > 0 && via[sink] == -1 {
			u := queue[0]
			queue = queue[1:]
			x.visit(u)
			for _, i := range adj[u] {
				if a := arcs[i]; a.cap > 0 && via[a.to] == -1 {
					via[a.to] = i
					queue = append(queue, a.to)
				}
			}
		}
		if via[sink] == -1 {
			break
		}
		push := Infinity
		path := []int{sink}
		for v := sink; v != source; v = arcs[via[v]^1].to {
			push = min(push, arcs[via[v]].cap)
			path = append(path, arcs[via[v]^1].to)
		}
		for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
			path[i], path[j] = path[j], path[i]
		}
		x.mark("path", path)
		for v := sink; v != source; v = arcs[via[v]^1].to {
			arcs[via[v]].cap -= push
			arcs[via[v]^1].cap += push
			e := via[v] / 2
			x.mark("flow", []int{e}, flowOf(e))
		}
		total += push
	}
	flow := make([]int, len(g.Edges))
	for i := range flow {
		flow[i] = flowOf(i)
	}
	return total, flow, nil
}
