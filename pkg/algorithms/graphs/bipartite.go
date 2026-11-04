package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Bipartite reports whether the nodes of g can be split into two sides with
// every edge between the sides, by coloring each component breadth-first. It
// returns each node's side (0 or 1) when it can, and otherwise an odd cycle,
// the proof that it cannot. Edge directions are ignored.
//
// Time: O(V + E). Space: O(V).
func Bipartite(g *Graph) (ok bool, side, oddCycle []int) { return bipartite(g, nop) }

// TraceBipartite is Bipartite, reporting each node colored (Mark "group",
// Values[0] the side) and an odd cycle (Mark "cycle").
func TraceBipartite(g *Graph, r trace.Recorder) (ok bool, side, oddCycle []int) {
	return bipartite(g, rec{r})
}

func bipartite(g *Graph, x rec) (bool, []int, []int) {
	und := *g
	und.Directed = false
	adj := und.adjacency()
	n := g.Len()
	side, parent := filled(n, -1), filled(n, -1)
	for s := 0; s < n; s++ {
		if side[s] >= 0 {
			continue
		}
		side[s] = 0
		x.mark("group", []int{s}, 0)
		queue := []int{s}
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			x.visit(u)
			for _, a := range adj[u] {
				v := a.to
				if side[v] < 0 {
					side[v], parent[v] = 1-side[u], u
					x.mark("group", []int{v}, side[v])
					queue = append(queue, v)
					continue
				}
				if side[v] == side[u] {
					c := oddCycle(parent, u, v)
					x.mark("cycle", c)
					return false, nil, c
				}
			}
		}
	}
	return true, side, nil
}

// oddCycle joins the tree paths from u and v back to their common ancestor;
// with the edge u-v that closes a cycle of odd length.
func oddCycle(parent []int, u, v int) []int {
	depth := func(a int) int {
		d := 0
		for ; parent[a] >= 0; a = parent[a] {
			d++
		}
		return d
	}
	var left, right []int
	du, dv := depth(u), depth(v)
	for ; du > dv; du-- {
		left, u = append(left, u), parent[u]
	}
	for ; dv > du; dv-- {
		right, v = append(right, v), parent[v]
	}
	for u != v {
		left, right = append(left, u), append(right, v)
		u, v = parent[u], parent[v]
	}
	left = append(left, u)
	for i := len(right) - 1; i >= 0; i-- {
		left = append(left, right[i])
	}
	return left
}
