package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Boruvka returns a minimum spanning forest of an undirected graph (edge
// indexes and total weight) with Boruvka's algorithm (1926), the oldest one:
// every component picks its cheapest edge to another component at once, all
// are added, and the round repeats; the components at least halve each round.
// Ties go to the lower edge index, so the forest is the same as Kruskal's
// when weights are distinct.
//
// Time: O(E log V). Space: O(V).
func Boruvka(g *Graph) (edges []int, total int, err error) { return boruvka(g, nop) }

// TraceBoruvka is Boruvka, reporting each round (Mark "range", Values[0] the
// round), each component's cheapest edge (Mark "edge") and each edge added
// (Mark "tree").
func TraceBoruvka(g *Graph, r trace.Recorder) (edges []int, total int, err error) {
	return boruvka(g, rec{r})
}

func boruvka(g *Graph, x rec) ([]int, int, error) {
	if g.Directed {
		return nil, 0, ErrDirected
	}
	d := newDisjoint(g.Len())
	var tree []int
	total := 0
	better := func(a, b int) bool { // is edge a cheaper than edge b (-1 none)?
		if b < 0 {
			return true
		}
		wa, wb := g.Edges[a].Weight, g.Edges[b].Weight
		return wa < wb || (wa == wb && a < b)
	}
	for round := 1; ; round++ {
		x.mark("range", nil, round)
		cheapest := filled(g.Len(), -1)
		for i, e := range g.Edges {
			a, b := d.find(e.From), d.find(e.To)
			if a == b {
				continue
			}
			if better(i, cheapest[a]) {
				cheapest[a] = i
			}
			if better(i, cheapest[b]) {
				cheapest[b] = i
			}
		}
		added := false
		for _, i := range cheapest {
			if i < 0 {
				continue
			}
			x.mark("edge", []int{i})
			if e := g.Edges[i]; d.union(e.From, e.To) {
				tree = append(tree, i)
				total += e.Weight
				x.mark("tree", []int{i})
				added = true
			}
		}
		if !added {
			return tree, total, nil
		}
	}
}
