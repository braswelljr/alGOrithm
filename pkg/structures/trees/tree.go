// Package trees is about binary trees: building them (from a level-order list,
// from sorted values, from two traversals), walking them (preorder, inorder,
// postorder, level order, and Morris inorder in O(1) space), measuring them
// (height, diameter, balance, the binary-search-tree property), finding the
// lowest common ancestor, serializing, summing paths, mirroring, and
// expression trees. One file each.
//
// A Tree is written as its level order with "null" for a missing child,
// "1 2 3 null 4" meaning 1 with children 2 and 3, and 4 the right child of 2,
// the form LeetCode uses. Every walk X has a TraceX that reports each node
// visited (trace.Visit with Values[0] the node's ID, its level-order position),
// and Snapshot draws the tree.
//
//	t, _ := trees.Parse("4 2 6 1 3 5 7")
//	trees.Inorder(t) // [1 2 3 4 5 6 7]
package trees

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MaxNodes bounds the trees Parse builds.
const MaxNodes = 10_000

// ErrInput is returned for text that is not a level-order tree.
var ErrInput = errors.New("trees: invalid tree")

// Node is a node of a binary tree. ID is its position in level order (the
// root is 0), set by Parse and Renumber.
type Node struct {
	Value       int
	Left, Right *Node
	ID          int
}

// Tree is a binary tree. The zero Tree is empty.
type Tree struct {
	Root *Node
}

// Parse reads a tree in level order: values separated by spaces or commas,
// "null" (or "nil", "#", "_") for a missing child, brackets allowed.
//
// Time: O(n). Space: O(n).
func Parse(text string) (*Tree, error) {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == ',' || r == '\n' || r == '\t' || r == '[' || r == ']' || r == '\r'
	})
	if len(fields) > MaxNodes {
		return nil, fmt.Errorf("%w: more than %d values", ErrInput, MaxNodes)
	}
	node := func(s string) (*Node, error) {
		switch strings.ToLower(s) {
		case "null", "nil", "#", "_":
			return nil, nil
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("%w: %q is not a number or null", ErrInput, s)
		}
		return &Node{Value: v}, nil
	}
	t := &Tree{}
	if len(fields) == 0 {
		return t, nil
	}
	root, err := node(fields[0])
	if err != nil {
		return nil, err
	}
	t.Root = root
	queue := []*Node{root}
	for i := 1; i < len(fields) && len(queue) > 0; {
		parent := queue[0]
		queue = queue[1:]
		if parent == nil {
			continue
		}
		for _, side := range []**Node{&parent.Left, &parent.Right} {
			if i >= len(fields) {
				break
			}
			n, err := node(fields[i])
			if err != nil {
				return nil, err
			}
			*side = n
			if n != nil {
				queue = append(queue, n)
			}
			i++
		}
	}
	t.Renumber()
	return t, nil
}

// FromSorted builds a balanced binary search tree from sorted values by
// making the middle value the root, recursively. Its height is the least
// possible, ceil(log2(n+1)).
//
// Time: O(n). Space: O(log n) recursion.
func FromSorted(sorted []int) *Tree {
	var build func(lo, hi int) *Node
	build = func(lo, hi int) *Node {
		if lo > hi {
			return nil
		}
		mid := lo + (hi-lo)/2
		return &Node{Value: sorted[mid], Left: build(lo, mid-1), Right: build(mid+1, hi)}
	}
	t := &Tree{Root: build(0, len(sorted)-1)}
	t.Renumber()
	return t
}

// Renumber sets every node's ID to its level-order position.
func (t *Tree) Renumber() {
	id := 0
	walkLevels(t.Root, func(n *Node) { n.ID = id; id++ })
}

// Size is the number of nodes. Time: O(n).
func (t *Tree) Size() int {
	n := 0
	walkLevels(t.Root, func(*Node) { n++ })
	return n
}

// Find returns the first node, in level order, holding v.
func (t *Tree) Find(v int) *Node {
	var found *Node
	walkLevels(t.Root, func(n *Node) {
		if found == nil && n.Value == v {
			found = n
		}
	})
	return found
}

// String writes the tree in the level order Parse reads, without trailing
// nulls.
func (t *Tree) String() string { return Serialize(t) }

// walkLevels calls f on every node in level order.
func walkLevels(root *Node, f func(*Node)) {
	if root == nil {
		return
	}
	queue := []*Node{root}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		f(n)
		for _, c := range []*Node{n.Left, n.Right} {
			if c != nil {
				queue = append(queue, c)
			}
		}
	}
}

// Snapshot describes the tree: nodes by ID with their values, edges labeled
// "left" and "right", the root's role "root". It implements trace.Snapshotter.
func (t *Tree) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	var walk func(n *Node, parent int, side string)
	walk = func(n *Node, parent int, side string) {
		if n == nil {
			return
		}
		nd := trace.Node{ID: n.ID, Label: strconv.Itoa(n.Value), Parent: parent}
		if parent < 0 {
			nd.Role = "root"
		} else {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: n.ID, Label: side})
		}
		s.Nodes = append(s.Nodes, nd)
		walk(n.Left, n.ID, "left")
		walk(n.Right, n.ID, "right")
	}
	walk(t.Root, -1, "")
	return s
}

func visit(r trace.Recorder, n *Node) {
	r.Record(trace.Event{Kind: trace.Visit, Values: []int{n.ID}})
}
