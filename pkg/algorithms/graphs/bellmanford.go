package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// BellmanFord returns the shortest distance from source to every node
// (Infinity where unreachable) and each node's predecessor, relaxing every
// edge V-1 times (Bellman and Ford, 1950s). Unlike Dijkstra it allows negative
// weights. If a negative cycle is reachable from source it returns
// ErrNegativeCycle and the cycle's nodes.
//
// Time: O(V * E). Space: O(V).
func BellmanFord(g *Graph, source int) (dist, prev, cycle []int, err error) {
	return bellmanFord(g, source, nop)
}

// TraceBellmanFord is BellmanFord, reporting each distance lowered (Set) and a
// negative cycle (Mark "cycle").
func TraceBellmanFord(g *Graph, source int, r trace.Recorder) (dist, prev, cycle []int, err error) {
	return bellmanFord(g, source, rec{r})
}

func bellmanFord(g *Graph, source int, x rec) ([]int, []int, []int, error) {
	if err := g.checkNode(source); err != nil {
		return nil, nil, nil, err
	}
	n := g.Len()
	dist, prev := filled(n, Infinity), filled(n, -1)
	dist[source] = 0
	x.set(source, 0)
	type relax struct{ from, to, w int }
	var arcs []relax
	for _, e := range g.Edges {
		arcs = append(arcs, relax{e.From, e.To, e.Weight})
		if !g.Directed {
			arcs = append(arcs, relax{e.To, e.From, e.Weight})
		}
	}
	last := -1
	for round := 0; round < n; round++ {
		last = -1
		for _, a := range arcs {
			if dist[a.from] != Infinity && dist[a.from]+a.w < dist[a.to] {
				dist[a.to], prev[a.to] = dist[a.from]+a.w, a.from
				x.set(a.to, dist[a.to])
				last = a.to
			}
		}
		if last < 0 {
			return dist, prev, nil, nil // nothing changed: done early
		}
	}
	// Still relaxing after n rounds: last is on, or leads to, a negative
	// cycle. Walk back n steps to land inside it, then collect it.
	at := last
	for i := 0; i < n; i++ {
		at = prev[at]
	}
	cycle := []int{at}
	for v := prev[at]; v != at; v = prev[v] {
		cycle = append(cycle, v)
	}
	for i, j := 0, len(cycle)-1; i < j; i, j = i+1, j-1 {
		cycle[i], cycle[j] = cycle[j], cycle[i]
	}
	x.mark("cycle", cycle)
	return dist, prev, cycle, ErrNegativeCycle
}
