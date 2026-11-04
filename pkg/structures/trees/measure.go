package trees

import "math"

// Height is the number of nodes on the longest root-to-leaf path; 0 for an
// empty tree.
//
// Time: O(n). Space: O(h).
func Height(t *Tree) int { return height(t.Root) }

func height(n *Node) int {
	if n == nil {
		return 0
	}
	return 1 + max(height(n.Left), height(n.Right))
}

// Diameter is the number of edges on the longest path between any two nodes,
// which need not pass through the root: at each node, the left height plus
// the right height.
//
// Time: O(n). Space: O(h).
func Diameter(t *Tree) int {
	best := 0
	var h func(*Node) int
	h = func(n *Node) int {
		if n == nil {
			return 0
		}
		l, r := h(n.Left), h(n.Right)
		best = max(best, l+r)
		return 1 + max(l, r)
	}
	h(t.Root)
	return best
}

// IsBalanced reports whether every node's subtrees differ in height by at
// most one (the AVL condition), in one bottom-up pass.
//
// Time: O(n). Space: O(h).
func IsBalanced(t *Tree) bool {
	var h func(*Node) int // -1 means unbalanced below
	h = func(n *Node) int {
		if n == nil {
			return 0
		}
		l, r := h(n.Left), h(n.Right)
		if l < 0 || r < 0 || l-r > 1 || r-l > 1 {
			return -1
		}
		return 1 + max(l, r)
	}
	return h(t.Root) >= 0
}

// IsBST reports whether the tree is a binary search tree with distinct
// values: every node greater than all of its left subtree and less than all
// of its right, checked by passing down the allowed range (checking only each
// node's children is the classic mistake).
//
// Time: O(n). Space: O(h).
func IsBST(t *Tree) bool {
	var ok func(n *Node, lo, hi int64) bool
	ok = func(n *Node, lo, hi int64) bool {
		if n == nil {
			return true
		}
		v := int64(n.Value)
		return v > lo && v < hi && ok(n.Left, lo, v) && ok(n.Right, v, hi)
	}
	return ok(t.Root, math.MinInt64, math.MaxInt64)
}
