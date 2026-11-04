package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// CycleDetect returns a cycle of g as its nodes, first node not repeated, or
// nil when g has none, by depth-first search: in a directed graph a cycle is
// an edge back to a node still on the search path; in an undirected graph,
// an edge to any visited node other than the one just come from (a self-loop
// or two parallel edges count).
//
// Time: O(V + E). Space: O(V).
func CycleDetect(g *Graph) []int { return cycleDetect(g, nop) }

// TraceCycleDetect is CycleDetect, reporting each node visited (Visit) and the
// cycle found (Mark "cycle").
func TraceCycleDetect(g *Graph, r trace.Recorder) []int { return cycleDetect(g, rec{r}) }

func cycleDetect(g *Graph, x rec) []int {
	const white, grey, black = 0, 1, 2
	adj := g.adjacency()
	color := make([]int, g.Len())
	parent := filled(g.Len(), -1)
	var found []int
	var visit func(u, viaEdge int) bool
	visit = func(u, viaEdge int) bool {
		color[u] = grey
		x.visit(u)
		for _, a := range adj[u] {
			if !g.Directed && a.edge == viaEdge {
				continue
			}
			v := a.to
			if color[v] == white {
				parent[v] = u
				if visit(v, a.edge) {
					return true
				}
				continue
			}
			if color[v] == grey || (!g.Directed && color[v] != white) {
				found = []int{v}
				for w := u; w != v && w >= 0; w = parent[w] {
					found = append(found, w)
				}
				// found runs v, u, parent(u), ...: reverse to follow the edges.
				for i, j := 1, len(found)-1; i < j; i, j = i+1, j-1 {
					found[i], found[j] = found[j], found[i]
				}
				x.mark("cycle", found)
				return true
			}
		}
		color[u] = black
		return false
	}
	for u := range color {
		if color[u] == white && visit(u, -1) {
			return found
		}
	}
	return nil
}
