package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// BreadthFirst returns the nodes reachable from start in breadth-first order:
// start, then every node one edge away, then two, each node once, neighbors in
// edge order.
//
// Time: O(V + E). Space: O(V).
func BreadthFirst(g *Graph, start int) ([]int, error) { return breadthFirst(g, start, nop) }

// TraceBreadthFirst is BreadthFirst, reporting each node queued (Push) and
// visited (Visit), and the order (Mark "order").
func TraceBreadthFirst(g *Graph, start int, r trace.Recorder) ([]int, error) {
	return breadthFirst(g, start, rec{r})
}

func breadthFirst(g *Graph, start int, x rec) ([]int, error) {
	if err := g.checkNode(start); err != nil {
		return nil, err
	}
	adj := g.adjacency()
	seen := make([]bool, g.Len())
	seen[start] = true
	queue := []int{start}
	x.push(start)
	var order []int
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		x.visit(u)
		order = append(order, u)
		for _, a := range adj[u] {
			if !seen[a.to] {
				seen[a.to] = true
				queue = append(queue, a.to)
				x.push(a.to)
			}
		}
	}
	x.mark("order", order)
	return order, nil
}

// DepthFirst returns the nodes reachable from start in depth-first preorder:
// follow the first unvisited neighbor as deep as it goes, then back up.
//
// Time: O(V + E). Space: O(V) for the stack.
func DepthFirst(g *Graph, start int) ([]int, error) { return depthFirst(g, start, nop) }

// TraceDepthFirst is DepthFirst, reporting each node pushed (Push) and visited
// (Visit), and the order (Mark "order").
func TraceDepthFirst(g *Graph, start int, r trace.Recorder) ([]int, error) {
	return depthFirst(g, start, rec{r})
}

func depthFirst(g *Graph, start int, x rec) ([]int, error) {
	if err := g.checkNode(start); err != nil {
		return nil, err
	}
	adj := g.adjacency()
	seen := make([]bool, g.Len())
	stack := []int{start}
	x.push(start)
	var order []int
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[u] {
			continue
		}
		seen[u] = true
		x.visit(u)
		order = append(order, u)
		for i := len(adj[u]) - 1; i >= 0; i-- { // reversed: the first neighbor is explored first
			if v := adj[u][i].to; !seen[v] {
				stack = append(stack, v)
				x.push(v)
			}
		}
	}
	x.mark("order", order)
	return order, nil
}
