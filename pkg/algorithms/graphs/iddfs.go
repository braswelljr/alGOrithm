package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// IterativeDeepening finds a path from start to target with the fewest edges
// by depth-first searches limited to depth 0, 1, 2, ...: the memory of
// depth-first search with the shortest-path guarantee of breadth-first
// search, at the cost of re-exploring the shallow levels. maxDepth of 0 or
// less means the number of nodes. It returns nil when target is not within
// maxDepth edges. Weights are ignored; a node is never repeated on the path.
//
// Time: O(b^d) for branching factor b and depth d. Space: O(d).
func IterativeDeepening(g *Graph, start, target, maxDepth int) ([]int, error) {
	return iddfs(g, start, target, maxDepth, nop)
}

// TraceIterativeDeepening is IterativeDeepening, reporting each new depth
// limit (Mark "range", Values[0] the limit), each node expanded (Visit), and
// the path (Mark "path").
func TraceIterativeDeepening(g *Graph, start, target, maxDepth int, r trace.Recorder) ([]int, error) {
	return iddfs(g, start, target, maxDepth, rec{r})
}

func iddfs(g *Graph, start, target, maxDepth int, x rec) ([]int, error) {
	for _, n := range []int{start, target} {
		if err := g.checkNode(n); err != nil {
			return nil, err
		}
	}
	if maxDepth <= 0 {
		maxDepth = g.Len()
	}
	adj := g.adjacency()
	onPath := make([]bool, g.Len())
	var path []int
	var dls func(u, depth int) bool
	dls = func(u, depth int) bool {
		x.visit(u)
		path = append(path, u)
		onPath[u] = true
		if u == target {
			return true
		}
		if depth > 0 {
			for _, a := range adj[u] {
				if !onPath[a.to] && dls(a.to, depth-1) {
					return true
				}
			}
		}
		onPath[u] = false
		path = path[:len(path)-1]
		return false
	}
	for depth := 0; depth <= maxDepth; depth++ {
		x.mark("range", nil, depth)
		if dls(start, depth) {
			x.mark("path", path)
			return path, nil
		}
	}
	return nil, nil
}
