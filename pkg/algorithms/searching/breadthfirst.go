package searching

// Node is a vertex of a graph: a value and the nodes it has edges to.
type Node[T comparable] struct {
	Value     T
	Neighbors []*Node[T]
}

// BreadthFirst returns the first node reachable from root whose Value is
// target, visiting nodes in order of their distance from root (level by
// level), or nil when there is none. Each node is visited once, so cycles are
// safe. The graph is not modified.
//
// Time: O(V + E) for the V nodes and E edges reachable from root.
// Space: O(V).
func BreadthFirst[T comparable](root *Node[T], target T) *Node[T] {
	return breadthFirst(root, target, nil)
}

// TraceBreadthFirst is BreadthFirst, calling visit with each node as it is
// visited, in order.
func TraceBreadthFirst[T comparable](root *Node[T], target T, visit func(*Node[T])) *Node[T] {
	return breadthFirst(root, target, visit)
}

func breadthFirst[T comparable](root *Node[T], target T, visit func(*Node[T])) *Node[T] {
	if root == nil {
		return nil
	}
	seen := map[*Node[T]]bool{root: true}
	queue := []*Node[T]{root}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if visit != nil {
			visit(n)
		}
		if n.Value == target {
			return n
		}
		for _, next := range n.Neighbors {
			if next != nil && !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return nil
}
