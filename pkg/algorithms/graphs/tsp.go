package graphs

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// weights returns the cheapest edge weight and its index between every
// ordered pair of nodes (Infinity and -1 when there is none).
func (g *Graph) weights() (w, edge [][]int) {
	n := g.Len()
	w, edge = make([][]int, n), make([][]int, n)
	for i := range w {
		w[i], edge[i] = filled(n, Infinity), filled(n, -1)
	}
	set := func(a, b, wt, i int) {
		if wt < w[a][b] {
			w[a][b], edge[a][b] = wt, i
		}
	}
	for i, e := range g.Edges {
		set(e.From, e.To, e.Weight, i)
		if !g.Directed {
			set(e.To, e.From, e.Weight, i)
		}
	}
	return w, edge
}

// HeldKarp returns the cheapest tour that starts at node 0, visits every node
// once and returns, with the Held-Karp dynamic program (1962): the cheapest
// way to reach node j having visited exactly the set S is built from the
// cheapest ways to reach S's other nodes. It is exact and exponential, so it
// accepts at most MaxHeldKarp nodes. The tour lists node 0 at both ends.
// Missing edges cannot be used; ErrNoTour when no tour exists.
//
// Time: O(2^n * n^2). Space: O(2^n * n).
func HeldKarp(g *Graph) (tour []int, cost int, err error) { return heldKarp(g, nop) }

// TraceHeldKarp is HeldKarp, reporting the tour's edges (Mark "tree") and
// nodes (Mark "path").
func TraceHeldKarp(g *Graph, r trace.Recorder) (tour []int, cost int, err error) {
	return heldKarp(g, rec{r})
}

func heldKarp(g *Graph, x rec) ([]int, int, error) {
	n := g.Len()
	if n > MaxHeldKarp {
		return nil, 0, fmt.Errorf("%w: at most %d", ErrTooManyNodes, MaxHeldKarp)
	}
	if n == 1 {
		return []int{0}, 0, nil
	}
	w, edge := g.weights()
	full := 1 << n
	best := make([][]int, full)
	from := make([][]int, full)
	for s := range best {
		best[s], from[s] = filled(n, Infinity), filled(n, -1)
	}
	best[1][0] = 0
	for s := 1; s < full; s += 2 { // every set holds node 0
		for j := 0; j < n; j++ {
			if s&(1<<j) == 0 || best[s][j] == Infinity {
				continue
			}
			for k := 0; k < n; k++ {
				if s&(1<<k) != 0 || w[j][k] == Infinity {
					continue
				}
				t := s | 1<<k
				if c := best[s][j] + w[j][k]; c < best[t][k] {
					best[t][k], from[t][k] = c, j
				}
			}
		}
	}
	cost, last := Infinity, -1
	for j := 1; j < n; j++ {
		if best[full-1][j] != Infinity && w[j][0] != Infinity && best[full-1][j]+w[j][0] < cost {
			cost, last = best[full-1][j]+w[j][0], j
		}
	}
	if last < 0 {
		return nil, 0, ErrNoTour
	}
	tour := []int{0}
	for s, j := full-1, last; j > 0; {
		tour = append(tour, j)
		s, j = s&^(1<<j), from[s][j]
	}
	tour = append(tour, 0)
	tour = reversed(tour)
	for i := 1; i < len(tour); i++ {
		x.mark("tree", []int{edge[tour[i-1]][tour[i]]})
	}
	x.mark("path", tour)
	return tour, cost, nil
}

// NearestNeighborTour builds a tour from start by always moving to the
// nearest unvisited node, then returning to start: fast and often within 25%
// of the best tour, but with no guarantee (compare HeldKarp). ErrNoTour when
// it gets stuck.
//
// Time: O(n^2). Space: O(n^2) for the weight table.
func NearestNeighborTour(g *Graph, start int) (tour []int, cost int, err error) {
	return nearestNeighbor(g, start, nop)
}

// TraceNearestNeighborTour is NearestNeighborTour, reporting each node moved
// to (Visit), each edge taken (Mark "tree") and the tour (Mark "path").
func TraceNearestNeighborTour(g *Graph, start int, r trace.Recorder) (tour []int, cost int, err error) {
	return nearestNeighbor(g, start, rec{r})
}

func nearestNeighbor(g *Graph, start int, x rec) ([]int, int, error) {
	if err := g.checkNode(start); err != nil {
		return nil, 0, err
	}
	w, edge := g.weights()
	n := g.Len()
	seen := make([]bool, n)
	seen[start] = true
	tour, cost := []int{start}, 0
	x.visit(start)
	for u := start; len(tour) < n; {
		next := -1
		for v := 0; v < n; v++ {
			if !seen[v] && w[u][v] != Infinity && (next < 0 || w[u][v] < w[u][next]) {
				next = v
			}
		}
		if next < 0 {
			return nil, 0, ErrNoTour
		}
		cost += w[u][next]
		x.mark("tree", []int{edge[u][next]})
		x.visit(next)
		seen[next], u = true, next
		tour = append(tour, next)
	}
	last := tour[len(tour)-1]
	if n > 1 {
		if w[last][start] == Infinity {
			return nil, 0, ErrNoTour
		}
		cost += w[last][start]
		x.mark("tree", []int{edge[last][start]})
		tour = append(tour, start)
	}
	x.mark("path", tour)
	return tour, cost, nil
}
