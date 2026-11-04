package trees

import (
	"fmt"
	"strconv"
	"strings"
)

// BuildFromPreIn rebuilds a tree from its preorder and inorder values, which
// must be distinct: the first preorder value is the root, and its place in the
// inorder list splits the rest into the left and right subtrees.
//
// Time: O(n) with an index of the inorder positions. Space: O(n).
func BuildFromPreIn(pre, in []int) (*Tree, error) {
	if len(pre) != len(in) {
		return nil, fmt.Errorf("%w: the traversals have different lengths", ErrInput)
	}
	pos := make(map[int]int, len(in))
	for i, v := range in {
		if _, dup := pos[v]; dup {
			return nil, fmt.Errorf("%w: %d appears twice; the values must be distinct", ErrInput, v)
		}
		pos[v] = i
	}
	next := 0
	var build func(lo, hi int) (*Node, error)
	build = func(lo, hi int) (*Node, error) {
		if lo > hi {
			return nil, nil
		}
		v := pre[next]
		next++
		i, ok := pos[v]
		if !ok || i < lo || i > hi {
			return nil, fmt.Errorf("%w: the traversals do not describe the same tree", ErrInput)
		}
		n := &Node{Value: v}
		var err error
		if n.Left, err = build(lo, i-1); err != nil {
			return nil, err
		}
		if n.Right, err = build(i+1, hi); err != nil {
			return nil, err
		}
		return n, nil
	}
	root, err := build(0, len(in)-1)
	if err != nil {
		return nil, err
	}
	t := &Tree{Root: root}
	t.Renumber()
	return t, nil
}

// Serialize writes the tree in level order with "null" for missing children,
// dropping trailing nulls, the form Parse reads back (Parse is Deserialize).
//
// Time: O(n). Space: O(n).
func Serialize(t *Tree) string {
	if t.Root == nil {
		return ""
	}
	var out []string
	queue := []*Node{t.Root}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if n == nil {
			out = append(out, "null")
			continue
		}
		out = append(out, strconv.Itoa(n.Value))
		queue = append(queue, n.Left, n.Right)
	}
	for len(out) > 0 && out[len(out)-1] == "null" {
		out = out[:len(out)-1]
	}
	return strings.Join(out, " ")
}

// PathSum returns every root-to-leaf path whose values sum to target, by
// depth-first search keeping the current path.
//
// Time: O(n * h) to copy the paths found. Space: O(h).
func PathSum(t *Tree, target int) [][]int {
	var out [][]int
	var path []int
	var walk func(n *Node, left int)
	walk = func(n *Node, left int) {
		if n == nil {
			return
		}
		path = append(path, n.Value)
		left -= n.Value
		if n.Left == nil && n.Right == nil && left == 0 {
			out = append(out, append([]int(nil), path...))
		}
		walk(n.Left, left)
		walk(n.Right, left)
		path = path[:len(path)-1]
	}
	walk(t.Root, target)
	return out
}

// Invert mirrors the tree in place, swapping every node's children, and
// returns it. Node IDs are renumbered to the new level order.
//
// Time: O(n). Space: O(h).
func Invert(t *Tree) *Tree {
	var inv func(*Node)
	inv = func(n *Node) {
		if n == nil {
			return
		}
		n.Left, n.Right = n.Right, n.Left
		inv(n.Left)
		inv(n.Right)
	}
	inv(t.Root)
	t.Renumber()
	return t
}
