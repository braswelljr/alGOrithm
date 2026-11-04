package graphs

import (
	"math"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// PageRank returns each node's PageRank (Brin and Page, 1998), the share of
// time a random surfer spends there who follows a random edge out with
// probability damping and jumps to a random node otherwise; ranks sum to 1.
// A node with no edges out shares its rank among all nodes. Weights are
// ignored and undirected edges count both ways. It iterates until the ranks
// move less than tolerance in total, or iterations times, and reports how
// many iterations it ran.
//
// Time: O(k * (V + E)) for k iterations. Space: O(V).
func PageRank(g *Graph, damping float64, iterations int, tolerance float64) (rank []float64, ran int) {
	return pageRank(g, damping, iterations, tolerance, nop)
}

// TracePageRank is PageRank, reporting every node's rank after each
// iteration (Set, Values[0] the rank in thousandths) and each iteration (Mark
// "range", Values[0] the iteration).
func TracePageRank(g *Graph, damping float64, iterations int, tolerance float64, r trace.Recorder) (rank []float64, ran int) {
	return pageRank(g, damping, iterations, tolerance, rec{r})
}

func pageRank(g *Graph, damping float64, iterations int, tolerance float64, x rec) ([]float64, int) {
	n := g.Len()
	adj := g.adjacency()
	rank := make([]float64, n)
	for i := range rank {
		rank[i] = 1 / float64(n)
	}
	ran := 0
	for ran < iterations {
		ran++
		x.mark("range", nil, ran)
		next := make([]float64, n)
		dangling := 0.0
		for u, out := range adj {
			if len(out) == 0 {
				dangling += rank[u]
				continue
			}
			share := rank[u] / float64(len(out))
			for _, a := range out {
				next[a.to] += share
			}
		}
		moved := 0.0
		for v := range next {
			next[v] = (1-damping)/float64(n) + damping*(next[v]+dangling/float64(n))
			moved += math.Abs(next[v] - rank[v])
			x.set(v, int(math.Round(next[v]*1000)))
		}
		rank = next
		if moved < tolerance {
			break
		}
	}
	return rank, ran
}
