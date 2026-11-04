package graphs

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// StoerWagner returns the global minimum cut of an undirected graph, the
// least total weight of edges whose removal splits it in two, and the nodes
// on one side, with the Stoer-Wagner algorithm (1997): each phase adds nodes
// in order of their total weight to those already added; the last two give a
// cut, and are merged; the best phase cut wins. Weights must not be negative;
// a disconnected graph has a cut of 0.
//
// Time: O(V^3) with a weight matrix. Space: O(V^2).
func StoerWagner(g *Graph) (cut int, side []int, err error) { return stoerWagner(g, nop) }

// TraceStoerWagner is StoerWagner, reporting each phase (Mark "range",
// Values[0] the phase), each node added in a phase (Visit), and the best cut's
// side (Mark "group", Values[0] 0 or 1, for every node).
func TraceStoerWagner(g *Graph, r trace.Recorder) (cut int, side []int, err error) {
	return stoerWagner(g, rec{r})
}

func stoerWagner(g *Graph, x rec) (int, []int, error) {
	if g.Directed {
		return 0, nil, ErrDirected
	}
	n := g.Len()
	if n < 2 {
		return 0, nil, errNeedTwoNodes
	}
	w := make([][]int, n)
	for i := range w {
		w[i] = make([]int, n)
	}
	for _, e := range g.Edges {
		if e.Weight < 0 {
			return 0, nil, fmt.Errorf("%w: weights must not be negative", ErrInput)
		}
		if e.From != e.To {
			w[e.From][e.To] += e.Weight
			w[e.To][e.From] += e.Weight
		}
	}
	members := make([][]int, n) // the original nodes merged into each node
	for i := range members {
		members[i] = []int{i}
	}
	alive := make([]bool, n)
	for i := range alive {
		alive[i] = true
	}
	best, bestSide := Infinity, []int(nil)
	for phase := 1; phase < n; phase++ {
		x.mark("range", nil, phase)
		weight := make([]int, n)
		added := make([]bool, n)
		prev, last := -1, -1
		for step := 0; step < n-phase+1; step++ {
			pick := -1
			for v := 0; v < n; v++ {
				if alive[v] && !added[v] && (pick < 0 || weight[v] > weight[pick]) {
					pick = v
				}
			}
			added[pick] = true
			x.visit(pick)
			prev, last = last, pick
			for v := 0; v < n; v++ {
				weight[v] += w[pick][v]
			}
		}
		if weight[last] < best {
			best, bestSide = weight[last], append([]int(nil), members[last]...)
		}
		members[prev] = append(members[prev], members[last]...)
		alive[last] = false
		for v := 0; v < n; v++ {
			w[prev][v] += w[last][v]
			w[v][prev] = w[prev][v]
		}
		w[prev][prev] = 0
	}
	in := make([]bool, n)
	for _, v := range bestSide {
		in[v] = true
	}
	for v := 0; v < n; v++ {
		s := 0
		if in[v] {
			s = 1
		}
		x.mark("group", []int{v}, s)
	}
	sort.Ints(bestSide)
	return best, bestSide, nil
}
