// Package searchtrees holds ordered structures that find, insert and remove
// a value in about log n steps by keeping it sorted as a tree: BST (unbalanced,
// to show why balance matters), AVL, RedBlackTree (left-leaning), BTree,
// Treap, SplayTree, IntervalTree, and the SkipList, which reaches the same cost
// with coin flips and linked levels instead of a tree.
//
// Each has a Snapshot method that draws it as a trace.Snapshot.
package searchtrees

import "github.com/braswelljr/alGOrithm/pkg/trace"

// binarySnapshot draws a binary tree in pre-order as a trace.Snapshot, with
// edges labeled "left" and "right". The trees differ only in what they show
// on a node, so each passes children, which returns a node's two children,
// and describe, which returns its label and role; nil is the zero N.
func binarySnapshot[N comparable](root N, children func(N) (left, right N), describe func(n N, isRoot bool) (label, role string)) trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	var none N
	var walk func(n N, parent int, side string)
	walk = func(n N, parent int, side string) {
		if n == none {
			return
		}
		id := len(s.Nodes)
		label, role := describe(n, parent < 0)
		s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: label, Role: role, Parent: parent})
		if parent >= 0 {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id, Label: side})
		}
		left, right := children(n)
		walk(left, id, "left")
		walk(right, id, "right")
	}
	walk(root, -1, "")
	return s
}

// rootRole is "root" for the root and role for every other node.
func rootRole(isRoot bool, role string) string {
	if isRoot {
		return "root"
	}
	return role
}
