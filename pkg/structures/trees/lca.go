package trees

// LowestCommonAncestor returns the deepest node that has both the node
// holding a and the node holding b in its subtree (a node is in its own
// subtree); ok is false when either value is absent. It works on any binary
// tree by asking each subtree whether it holds a or b.
//
// Time: O(n). Space: O(h).
func LowestCommonAncestor(t *Tree, a, b int) (*Node, bool) {
	if t.Find(a) == nil || t.Find(b) == nil {
		return nil, false
	}
	var find func(n *Node) *Node
	find = func(n *Node) *Node {
		if n == nil || n.Value == a || n.Value == b {
			return n
		}
		l, r := find(n.Left), find(n.Right)
		switch {
		case l != nil && r != nil:
			return n // a on one side, b on the other
		case l != nil:
			return l
		}
		return r
	}
	return find(t.Root), true
}

// LowestCommonAncestorBST is LowestCommonAncestor for a binary search tree:
// walk down from the root while a and b are on the same side.
//
// Time: O(h). Space: O(1).
func LowestCommonAncestorBST(t *Tree, a, b int) (*Node, bool) {
	if t.Find(a) == nil || t.Find(b) == nil {
		return nil, false
	}
	n := t.Root
	for n != nil {
		switch {
		case a < n.Value && b < n.Value:
			n = n.Left
		case a > n.Value && b > n.Value:
			n = n.Right
		default:
			return n, true
		}
	}
	return nil, false
}
