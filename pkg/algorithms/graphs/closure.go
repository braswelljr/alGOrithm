package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// TransitiveClosure returns reach, with reach[i][j] 1 when j can be reached
// from i by one or more edges and 0 otherwise, with Warshall's algorithm
// (1962): j is reachable from i through k when k is reachable from i and j
// from k. On an undirected graph every node with an edge reaches itself (out
// and back).
//
// Time: O(V^3). Space: O(V^2).
func TransitiveClosure(g *Graph) [][]int { return closure(g, nop) }

// TraceTransitiveClosure is TransitiveClosure, reporting each pair found
// reachable (Set, Indices[0] i*V+j, Values[0] 1) and each intermediate node
// tried (Visit).
func TraceTransitiveClosure(g *Graph, r trace.Recorder) [][]int { return closure(g, rec{r}) }

func closure(g *Graph, x rec) [][]int {
	n := g.Len()
	reach := make([][]int, n)
	for i := range reach {
		reach[i] = make([]int, n)
	}
	mark := func(i, j int) {
		if reach[i][j] == 0 {
			reach[i][j] = 1
			x.r.Record(trace.Event{Kind: trace.Set, Indices: []int{i*n + j}, Values: []int{1}})
		}
	}
	for _, e := range g.Edges {
		mark(e.From, e.To)
		if !g.Directed {
			mark(e.To, e.From)
		}
	}
	for k := 0; k < n; k++ {
		x.visit(k)
		for i := 0; i < n; i++ {
			if reach[i][k] == 0 {
				continue
			}
			for j := 0; j < n; j++ {
				if reach[k][j] == 1 {
					mark(i, j)
				}
			}
		}
	}
	return reach
}
