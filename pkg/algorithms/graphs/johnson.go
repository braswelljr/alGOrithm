package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Johnson returns the shortest distance between every pair of nodes
// (Infinity where unreachable) with Johnson's algorithm (1977): Bellman-Ford
// from a new node joined to every node by a zero edge gives each node a
// potential h, reweighting every edge u-v to w + h(u) - h(v) makes every weight
// non-negative without changing which paths are shortest, and then Dijkstra
// runs from every node. It allows negative weights and returns
// ErrNegativeCycle when a cycle is negative; on sparse graphs it beats
// Floyd-Warshall.
//
// Time: O(V * E log V). Space: O(V^2) for the result.
func Johnson(g *Graph) ([][]int, error) { return johnson(g, nop) }

// TraceJohnson is Johnson, reporting each node's potential (Set, Values[0] h)
// and each source run from (Visit).
func TraceJohnson(g *Graph, r trace.Recorder) ([][]int, error) { return johnson(g, rec{r}) }

func johnson(g *Graph, x rec) ([][]int, error) {
	n := g.Len()
	arcs := make([]Edge, 0, 2*len(g.Edges))
	for _, e := range g.Edges {
		arcs = append(arcs, e)
		if !g.Directed {
			arcs = append(arcs, Edge{From: e.To, To: e.From, Weight: e.Weight})
		}
	}
	h := make([]int, n) // Bellman-Ford from a virtual node: every h starts at 0
	for round := 0; round <= n; round++ {
		changed := false
		for _, a := range arcs {
			if h[a.From]+a.Weight < h[a.To] {
				h[a.To] = h[a.From] + a.Weight
				changed = true
			}
		}
		if !changed {
			break
		}
		if round == n {
			return nil, ErrNegativeCycle
		}
	}
	for v, p := range h {
		x.set(v, p)
	}
	re := &Graph{Directed: true, Names: g.Names}
	for _, a := range arcs {
		re.Edges = append(re.Edges, Edge{From: a.From, To: a.To, Weight: a.Weight + h[a.From] - h[a.To]})
	}
	out := make([][]int, n)
	for s := 0; s < n; s++ {
		x.visit(s)
		d, _, err := dijkstra(re, s, nop)
		if err != nil {
			return nil, err
		}
		for v := range d {
			if d[v] != Infinity {
				d[v] += h[v] - h[s]
			}
		}
		out[s] = d
	}
	return out, nil
}
