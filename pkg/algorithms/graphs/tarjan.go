package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// TarjanSCC returns the strongly connected components of g, the largest sets
// of nodes that can each reach every other, with Tarjan's algorithm (1972):
// one depth-first search keeping each node's discovery index and the lowest
// index it can reach. Components come out in reverse topological order of the
// graph of components. On an undirected graph they are the connected
// components.
//
// Time: O(V + E). Space: O(V).
func TarjanSCC(g *Graph) [][]int { return tarjan(g, nop) }

// TraceTarjanSCC is TarjanSCC, reporting each node discovered (Visit) and
// pushed on the component stack (Push), and each component (Mark "group",
// Values[0] its number).
func TraceTarjanSCC(g *Graph, r trace.Recorder) [][]int { return tarjan(g, rec{r}) }

func tarjan(g *Graph, x rec) [][]int {
	adj := g.adjacency()
	n := g.Len()
	index, low := filled(n, -1), make([]int, n)
	onStack := make([]bool, n)
	var stack []int
	var comps [][]int
	next := 0
	var visit func(u int)
	visit = func(u int) {
		index[u], low[u] = next, next
		next++
		x.visit(u)
		stack = append(stack, u)
		onStack[u] = true
		x.push(u)
		// An undirected edge is two arcs, so its way back counts: each
		// connected component is then one strongly connected component.
		for _, a := range adj[u] {
			switch {
			case index[a.to] < 0:
				visit(a.to)
				low[u] = min(low[u], low[a.to])
			case onStack[a.to]:
				low[u] = min(low[u], index[a.to])
			}
		}
		if low[u] != index[u] {
			return
		}
		var comp []int
		for {
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[v] = false
			comp = append(comp, v)
			if v == u {
				break
			}
		}
		x.mark("group", comp, len(comps))
		comps = append(comps, comp)
	}
	for u := 0; u < n; u++ {
		if index[u] < 0 {
			visit(u)
		}
	}
	return comps
}
