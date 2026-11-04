package graphs

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// HopcroftKarp returns a maximum matching of a bipartite graph, the largest
// set of edges with no two sharing a node, as edge indexes in ascending
// order, with the Hopcroft-Karp algorithm (1973): breadth-first search finds
// the length of the shortest augmenting paths, then depth-first search
// augments along as many node-disjoint ones of that length as it can. The two
// sides are found by Bipartite (edge directions are ignored); a graph that is
// not bipartite returns ErrNotBipartite.
//
// Time: O(E * sqrt V). Space: O(V).
func HopcroftKarp(g *Graph) ([]int, error) { return hopcroftKarp(g, nop) }

// TraceHopcroftKarp is HopcroftKarp, reporting each augmenting path (Mark
// "path") and each matched edge at the end (Mark "tree").
func TraceHopcroftKarp(g *Graph, r trace.Recorder) ([]int, error) { return hopcroftKarp(g, rec{r}) }

func hopcroftKarp(g *Graph, x rec) ([]int, error) {
	ok, side, _ := bipartite(g, nop)
	if !ok {
		return nil, ErrNotBipartite
	}
	und := *g
	und.Directed = false
	adj := und.adjacency()
	n := g.Len()
	mate, mateEdge := filled(n, -1), filled(n, -1)
	dist := make([]int, n)
	bfs := func() bool {
		var queue []int
		found := false
		for u := 0; u < n; u++ {
			dist[u] = Infinity
			if side[u] == 0 && mate[u] < 0 {
				dist[u] = 0
				queue = append(queue, u)
			}
		}
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			for _, a := range adj[u] {
				w := mate[a.to]
				if w < 0 {
					found = true
				} else if dist[w] == Infinity {
					dist[w] = dist[u] + 1
					queue = append(queue, w)
				}
			}
		}
		return found
	}
	var dfs func(u int, path []int) bool
	dfs = func(u int, path []int) bool {
		for _, a := range adj[u] {
			w := mate[a.to]
			if w < 0 || (dist[w] == dist[u]+1 && dfs(w, append(path, a.to, w))) {
				if w < 0 {
					x.mark("path", append(path, a.to))
				}
				mate[u], mate[a.to] = a.to, u
				mateEdge[u], mateEdge[a.to] = a.edge, a.edge
				return true
			}
		}
		dist[u] = Infinity
		return false
	}
	for bfs() {
		for u := 0; u < n; u++ {
			if side[u] == 0 && mate[u] < 0 {
				dfs(u, []int{u})
			}
		}
	}
	var matched []int
	for u := 0; u < n; u++ {
		if side[u] == 0 && mate[u] >= 0 {
			matched = append(matched, mateEdge[u])
		}
	}
	sort.Ints(matched)
	for _, e := range matched {
		x.mark("tree", []int{e})
	}
	return matched, nil
}
