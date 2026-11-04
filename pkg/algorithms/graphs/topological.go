package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// TopologicalSort returns the nodes of a directed graph in an order where
// every edge goes from an earlier node to a later one, with Kahn's algorithm
// (1962): repeatedly take a node nothing points to and remove its edges. Ties
// go to the node that appeared first. A cycle returns ErrCycle and the order
// found before it got stuck.
//
// Time: O(V + E). Space: O(V).
func TopologicalSort(g *Graph) ([]int, error) { return topological(g, nop) }

// TraceTopologicalSort is TopologicalSort, reporting each node that becomes
// free (Push), each one taken (Visit), and the order (Mark "order").
func TraceTopologicalSort(g *Graph, r trace.Recorder) ([]int, error) {
	return topological(g, rec{r})
}

func topological(g *Graph, x rec) ([]int, error) {
	if !g.Directed {
		return nil, ErrUndirected
	}
	adj := g.adjacency()
	in := make([]int, g.Len())
	for _, e := range g.Edges {
		in[e.To]++
	}
	var ready, order []int
	for v, d := range in {
		if d == 0 {
			ready = append(ready, v)
			x.push(v)
		}
	}
	for len(ready) > 0 {
		// The lowest-numbered free node: a stable, predictable order.
		best := 0
		for i, v := range ready {
			if v < ready[best] {
				best = i
			}
		}
		u := ready[best]
		ready = append(ready[:best], ready[best+1:]...)
		x.visit(u)
		order = append(order, u)
		for _, a := range adj[u] {
			if in[a.to]--; in[a.to] == 0 {
				ready = append(ready, a.to)
				x.push(a.to)
			}
		}
	}
	x.mark("order", order)
	if len(order) < g.Len() {
		return order, ErrCycle
	}
	return order, nil
}
