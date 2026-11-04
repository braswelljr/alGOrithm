package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// FloydWarshall returns the shortest distance between every pair of nodes,
// dist[i][j] (Infinity where j cannot be reached from i): for each node k in
// turn, check whether going through k shortens any path (Floyd and Warshall,
// 1962). Negative weights are allowed; a negative cycle returns
// ErrNegativeCycle.
//
// Time: O(V^3). Space: O(V^2).
func FloydWarshall(g *Graph) ([][]int, error) { return floydWarshall(g, nop) }

// TraceFloydWarshall is FloydWarshall, reporting each improved entry (Set,
// Indices[0] is i*V+j) and each intermediate node (Visit).
func TraceFloydWarshall(g *Graph, r trace.Recorder) ([][]int, error) {
	return floydWarshall(g, rec{r})
}

func floydWarshall(g *Graph, x rec) ([][]int, error) {
	n := g.Len()
	d := make([][]int, n)
	for i := range d {
		d[i] = filled(n, Infinity)
		d[i][i] = 0
	}
	setMin := func(i, j, w int) {
		if w < d[i][j] {
			d[i][j] = w
		}
	}
	for _, e := range g.Edges {
		setMin(e.From, e.To, e.Weight)
		if !g.Directed {
			setMin(e.To, e.From, e.Weight)
		}
	}
	for k := 0; k < n; k++ {
		x.visit(k)
		for i := 0; i < n; i++ {
			if d[i][k] == Infinity {
				continue
			}
			for j := 0; j < n; j++ {
				if d[k][j] != Infinity && d[i][k]+d[k][j] < d[i][j] {
					d[i][j] = d[i][k] + d[k][j]
					x.r.Record(trace.Event{Kind: trace.Set, Indices: []int{i*n + j}, Values: []int{d[i][j]}})
				}
			}
		}
	}
	for i := range d {
		if d[i][i] < 0 {
			return d, ErrNegativeCycle
		}
	}
	return d, nil
}
