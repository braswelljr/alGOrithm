package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Kosaraju returns the strongly connected components of g with Kosaraju's
// algorithm (1978): a depth-first search records the order nodes finish, then
// a second search over the reversed edges, taking nodes latest-finished
// first, collects one component per tree. It finds the same components as
// TarjanSCC with two simple passes instead of one subtle one. On an
// undirected graph they are the connected components.
//
// Time: O(V + E). Space: O(V + E) for the reversed edges.
func Kosaraju(g *Graph) [][]int { return kosaraju(g, nop) }

// TraceKosaraju is Kosaraju, reporting each node finished in the first pass
// (Visit) and each node given a component in the second (Mark "group",
// Values[0] the component).
func TraceKosaraju(g *Graph, r trace.Recorder) [][]int { return kosaraju(g, rec{r}) }

func kosaraju(g *Graph, x rec) [][]int {
	adj, back := g.adjacency(), g.reverseAdjacency()
	seen := make([]bool, g.Len())
	var order []int
	var finish func(u int)
	finish = func(u int) {
		seen[u] = true
		for _, a := range adj[u] {
			if !seen[a.to] {
				finish(a.to)
			}
		}
		x.visit(u)
		order = append(order, u)
	}
	for u := range seen {
		if !seen[u] {
			finish(u)
		}
	}
	comp := filled(g.Len(), -1)
	var comps [][]int
	var collect func(u, c int)
	collect = func(u, c int) {
		comp[u] = c
		comps[c] = append(comps[c], u)
		x.mark("group", []int{u}, c)
		for _, a := range back[u] {
			if comp[a.to] < 0 {
				collect(a.to, c)
			}
		}
	}
	for i := len(order) - 1; i >= 0; i-- {
		if u := order[i]; comp[u] < 0 {
			comps = append(comps, nil)
			collect(u, len(comps)-1)
		}
	}
	return comps
}
