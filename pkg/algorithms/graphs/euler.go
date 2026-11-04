package graphs

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// EulerPath returns a walk that uses every edge exactly once, as its nodes and
// the edges in the order taken, with Hierholzer's algorithm (1873): walk until
// stuck, then splice in a detour from the latest node with unused edges. On an
// undirected graph one exists when every edge is connected and zero or two
// nodes have odd degree (it starts at one of them); on a directed graph when
// every node has as many edges in as out, except possibly one start with one
// more out and one end with one more in. Otherwise it returns ErrNoEuler with
// the reason. A graph with no edges has the one-node path of its first node.
//
// Time: O(V + E). Space: O(V + E).
func EulerPath(g *Graph) (nodes, edges []int, err error) { return euler(g, nop) }

// TraceEulerPath is EulerPath, reporting each node walked to (Visit), each
// edge used (Mark "tree", Indices[0] the edge), and the path (Mark "path").
func TraceEulerPath(g *Graph, r trace.Recorder) (nodes, edges []int, err error) {
	return euler(g, rec{r})
}

func euler(g *Graph, x rec) ([]int, []int, error) {
	if len(g.Edges) == 0 {
		return []int{0}, nil, nil
	}
	start, err := eulerStart(g)
	if err != nil {
		return nil, nil, err
	}
	adj := g.adjacency()
	used := make([]bool, len(g.Edges))
	next := make([]int, g.Len()) // next arc of each node to try
	type step struct{ node, edge int }
	stack := []step{{start, -1}}
	var walk []step
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		u := top.node
		for next[u] < len(adj[u]) && used[adj[u][next[u]].edge] {
			next[u]++
		}
		if next[u] == len(adj[u]) {
			walk = append(walk, top)
			stack = stack[:len(stack)-1]
			continue
		}
		a := adj[u][next[u]]
		used[a.edge] = true
		x.visit(a.to)
		x.mark("tree", []int{a.edge})
		stack = append(stack, step{a.to, a.edge})
	}
	if len(walk) != len(g.Edges)+1 {
		return nil, nil, fmt.Errorf("%w: the edges are not all connected", ErrNoEuler)
	}
	nodes := make([]int, 0, len(walk))
	edges := make([]int, 0, len(g.Edges))
	for i := len(walk) - 1; i >= 0; i-- {
		nodes = append(nodes, walk[i].node)
		if walk[i].edge >= 0 {
			edges = append(edges, walk[i].edge)
		}
	}
	x.mark("path", nodes)
	return nodes, edges, nil
}

// eulerStart checks the degree conditions and returns where the path starts.
func eulerStart(g *Graph) (int, error) {
	out, in := make([]int, g.Len()), make([]int, g.Len())
	for _, e := range g.Edges {
		out[e.From]++
		in[e.To]++
	}
	start := g.Edges[0].From
	if !g.Directed {
		var odd []int
		for v := range out {
			if (out[v]+in[v])%2 == 1 {
				odd = append(odd, v)
			}
		}
		switch len(odd) {
		case 0:
			return start, nil
		case 2:
			return odd[0], nil
		}
		return 0, fmt.Errorf("%w: %d nodes have odd degree (at most 2 may)", ErrNoEuler, len(odd))
	}
	starts, ends := 0, 0
	for v := range out {
		switch d := out[v] - in[v]; {
		case d == 1:
			starts++
			start = v
		case d == -1:
			ends++
		case d != 0:
			return 0, fmt.Errorf("%w: %s has %d edges out and %d in", ErrNoEuler, g.Names[v], out[v], in[v])
		}
	}
	if starts > 1 || ends > 1 || starts != ends {
		return 0, fmt.Errorf("%w: %d nodes have an extra edge out and %d an extra edge in (at most one each)", ErrNoEuler, starts, ends)
	}
	return start, nil
}
