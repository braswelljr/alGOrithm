package graphs

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Dinic returns the maximum flow from source to sink with Dinic's algorithm
// (1970), treating weights as capacities: build a level graph of shortest
// residual distances breadth-first, push a blocking flow through it
// depth-first, and repeat. It gives the same answer as MaxFlow (Edmonds-Karp)
// in fewer phases. The flow on each edge is as MaxFlow reports it.
//
// Time: O(V^2 * E). Space: O(V + E).
func Dinic(g *Graph, source, sink int) (total int, flow []int, err error) {
	return dinic(g, source, sink, nop)
}

// TraceDinic is Dinic, reporting each phase (Mark "range", Values[0] the
// phase), each node leveled (Visit), each augmenting path (Mark "path") and
// each edge's new flow (Mark "flow").
func TraceDinic(g *Graph, source, sink int, r trace.Recorder) (total int, flow []int, err error) {
	return dinic(g, source, sink, rec{r})
}

func dinic(g *Graph, source, sink int, x rec) (int, []int, error) {
	for _, n := range []int{source, sink} {
		if err := g.checkNode(n); err != nil {
			return 0, nil, err
		}
	}
	if source == sink {
		return 0, nil, fmt.Errorf("%w: the source is the sink", ErrInput)
	}
	type residual struct{ to, cap int }
	arcs := make([]residual, 0, 2*len(g.Edges))
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
		arcs = append(arcs, residual{e.To, e.Weight})
		adj[e.To] = append(adj[e.To], len(arcs))
		arcs = append(arcs, residual{e.From, back})
	}
	flowOf := func(i int) int { return g.Edges[i].Weight - arcs[2*i].cap }
	level := make([]int, g.Len())
	next := make([]int, g.Len())
	var push func(u, limit int, path []int) int
	push = func(u, limit int, path []int) int {
		if u == sink {
			x.mark("path", path)
			return limit
		}
		for ; next[u] < len(adj[u]); next[u]++ {
			i := adj[u][next[u]]
			a := arcs[i]
			if a.cap <= 0 || level[a.to] != level[u]+1 {
				continue
			}
			if f := push(a.to, min(limit, a.cap), append(path, a.to)); f > 0 {
				arcs[i].cap -= f
				arcs[i^1].cap += f
				x.mark("flow", []int{i / 2}, flowOf(i/2))
				return f
			}
		}
		return 0
	}
	total := 0
	for phase := 1; ; phase++ {
		for i := range level {
			level[i] = -1
		}
		level[source] = 0
		queue := []int{source}
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			x.visit(u)
			for _, i := range adj[u] {
				if a := arcs[i]; a.cap > 0 && level[a.to] < 0 {
					level[a.to] = level[u] + 1
					queue = append(queue, a.to)
				}
			}
		}
		if level[sink] < 0 {
			break
		}
		x.mark("range", nil, phase)
		for i := range next {
			next[i] = 0
		}
		for f := push(source, Infinity, []int{source}); f > 0; f = push(source, Infinity, []int{source}) {
			total += f
		}
	}
	flow := make([]int, len(g.Edges))
	for i := range flow {
		flow[i] = flowOf(i)
	}
	return total, flow, nil
}
