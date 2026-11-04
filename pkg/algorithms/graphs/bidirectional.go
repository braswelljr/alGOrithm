package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// BidirectionalBFS finds a path from start to target with the fewest edges by
// searching from both ends at once, forward from start and backward from
// target, one whole level at a time from the side with the smaller frontier,
// until they meet. Each side only explores about half the depth, so on a
// graph that branches b ways it visits about 2*b^(d/2) nodes instead of b^d.
// It returns nil when target cannot be reached. Weights are ignored.
//
// Time: O(b^(d/2)). Space: O(b^(d/2)).
func BidirectionalBFS(g *Graph, start, target int) ([]int, error) {
	return bidirectional(g, start, target, nop)
}

// TraceBidirectionalBFS is BidirectionalBFS, reporting each node reached
// (Mark "group", Values[0] 0 from the start side, 1 from the target side),
// each node expanded (Visit), and the path (Mark "path").
func TraceBidirectionalBFS(g *Graph, start, target int, r trace.Recorder) ([]int, error) {
	return bidirectional(g, start, target, rec{r})
}

func bidirectional(g *Graph, start, target int, x rec) ([]int, error) {
	for _, n := range []int{start, target} {
		if err := g.checkNode(n); err != nil {
			return nil, err
		}
	}
	if start == target {
		x.mark("path", []int{start})
		return []int{start}, nil
	}
	adj := [2][][]arc{g.adjacency(), g.reverseAdjacency()}
	dist := [2][]int{filled(g.Len(), -1), filled(g.Len(), -1)}
	parent := [2][]int{filled(g.Len(), -1), filled(g.Len(), -1)}
	frontier := [2][]int{{start}, {target}}
	dist[0][start], dist[1][target] = 0, 0
	x.mark("group", []int{start}, 0)
	x.mark("group", []int{target}, 1)
	meet, best := -1, Infinity
	for len(frontier[0]) > 0 && len(frontier[1]) > 0 {
		side := 0
		if len(frontier[1]) < len(frontier[0]) {
			side = 1
		}
		other := 1 - side
		var next []int
		for _, u := range frontier[side] {
			x.visit(u)
			for _, a := range adj[side][u] {
				v := a.to
				if dist[side][v] < 0 {
					dist[side][v], parent[side][v] = dist[side][u]+1, u
					x.mark("group", []int{v}, side)
					next = append(next, v)
				}
				if dist[other][v] >= 0 && dist[side][v]+dist[other][v] < best {
					best, meet = dist[side][v]+dist[other][v], v
				}
			}
		}
		frontier[side] = next
		if meet >= 0 { // the whole level is done, so no shorter meeting exists
			break
		}
	}
	if meet < 0 {
		return nil, nil
	}
	var path []int
	for v := meet; v >= 0; v = parent[0][v] {
		path = append(path, v)
	}
	path = reversed(path)
	for v := parent[1][meet]; v >= 0; v = parent[1][v] {
		path = append(path, v)
	}
	x.mark("path", path)
	return path, nil
}
