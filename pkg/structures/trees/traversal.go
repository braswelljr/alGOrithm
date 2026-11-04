package trees

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Preorder returns the values root first, then the left subtree, then the
// right: the order to copy a tree.
//
// Time: O(n). Space: O(h) for height h.
func Preorder(t *Tree) []int { return TracePreorder(t, trace.Nop{}) }

// TracePreorder is Preorder, reporting each node visited.
func TracePreorder(t *Tree, r trace.Recorder) []int {
	var out []int
	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		visit(r, n)
		out = append(out, n.Value)
		walk(n.Left)
		walk(n.Right)
	}
	walk(t.Root)
	return out
}

// Inorder returns the values left subtree first, then the root, then the
// right: ascending order for a binary search tree.
//
// Time: O(n). Space: O(h).
func Inorder(t *Tree) []int { return TraceInorder(t, trace.Nop{}) }

// TraceInorder is Inorder, reporting each node visited.
func TraceInorder(t *Tree, r trace.Recorder) []int {
	var out []int
	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		walk(n.Left)
		visit(r, n)
		out = append(out, n.Value)
		walk(n.Right)
	}
	walk(t.Root)
	return out
}

// Postorder returns the values both subtrees first, then the root: the order
// to delete a tree or evaluate an expression.
//
// Time: O(n). Space: O(h).
func Postorder(t *Tree) []int { return TracePostorder(t, trace.Nop{}) }

// TracePostorder is Postorder, reporting each node visited.
func TracePostorder(t *Tree, r trace.Recorder) []int {
	var out []int
	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		walk(n.Left)
		walk(n.Right)
		visit(r, n)
		out = append(out, n.Value)
	}
	walk(t.Root)
	return out
}

// LevelOrder returns the values level by level, left to right, with a queue
// (breadth-first search).
//
// Time: O(n). Space: O(w) for the widest level w.
func LevelOrder(t *Tree) [][]int { return TraceLevelOrder(t, trace.Nop{}) }

// TraceLevelOrder is LevelOrder, reporting each node visited.
func TraceLevelOrder(t *Tree, r trace.Recorder) [][]int {
	var out [][]int
	if t.Root == nil {
		return out
	}
	level := []*Node{t.Root}
	for len(level) > 0 {
		var vals []int
		var next []*Node
		for _, n := range level {
			visit(r, n)
			vals = append(vals, n.Value)
			for _, c := range []*Node{n.Left, n.Right} {
				if c != nil {
					next = append(next, c)
				}
			}
		}
		out = append(out, vals)
		level = next
	}
	return out
}

// Morris returns the inorder values in O(1) extra space (Joseph Morris,
// 1979): instead of a stack, it threads each node's inorder predecessor's
// empty right pointer back to the node, follows the threads, and removes them
// on the way back, leaving the tree as it found it.
//
// Time: O(n) (each edge is walked at most three times). Space: O(1).
func Morris(t *Tree) []int { return TraceMorris(t, trace.Nop{}) }

// TraceMorris is Morris, reporting each node when it is output (Visit) and
// each thread made or removed (Link and Unlink, Values [from ID, to ID]).
func TraceMorris(t *Tree, r trace.Recorder) []int {
	var out []int
	cur := t.Root
	for cur != nil {
		if cur.Left == nil {
			visit(r, cur)
			out = append(out, cur.Value)
			cur = cur.Right
			continue
		}
		pred := cur.Left
		for pred.Right != nil && pred.Right != cur {
			pred = pred.Right
		}
		if pred.Right == nil {
			pred.Right = cur // thread: come back to cur after the left subtree
			r.Record(trace.Event{Kind: trace.Link, Values: []int{pred.ID, cur.ID}})
			cur = cur.Left
			continue
		}
		pred.Right = nil // the left subtree is done: remove the thread
		r.Record(trace.Event{Kind: trace.Unlink, Values: []int{pred.ID, cur.ID}})
		visit(r, cur)
		out = append(out, cur.Value)
		cur = cur.Right
	}
	return out
}
