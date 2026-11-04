package buffers

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// leafSize is the longest string a Rope keeps in one leaf.
const leafSize = 8

type ropeNode struct {
	text        string // leaves only
	weight      int    // the length of the left subtree (a leaf: its text)
	left, right *ropeNode
}

func (n *ropeNode) length() int {
	if n == nil {
		return 0
	}
	if n.left == nil && n.right == nil {
		return len(n.text)
	}
	return n.weight + n.right.length()
}

// Rope is text held as a binary tree whose leaves are short strings (Hans
// Boehm, Russ Atkinson and Michael Plass, 1995). Each inner node stores the
// length of its left subtree, so finding position i walks one path. Split
// cuts the tree along a path and Concat joins two trees under a new root,
// neither copying the text, so inserting or deleting in the middle of a long
// document costs about log n, not n.
//
// Inserts here are not rebalanced, so after many edits the depth d can grow
// past log n; Rebalance rebuilds a balanced rope.
//
// It works on bytes; multi-byte UTF-8 characters count as several positions.
//
// Time: O(d) Index, Split, Insert and Delete; O(1) Concat; O(n) String.
// Space: O(n).
type Rope struct{ root *ropeNode }

// NewRope returns a balanced rope holding s.
func NewRope(s string) *Rope { return &Rope{build(s)} }

func build(s string) *ropeNode {
	if len(s) <= leafSize {
		if s == "" {
			return nil
		}
		return &ropeNode{text: s, weight: len(s)}
	}
	m := len(s) / 2
	return join(build(s[:m]), build(s[m:]))
}

func join(a, b *ropeNode) *ropeNode {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return &ropeNode{weight: a.length(), left: a, right: b}
}

// Len returns the length of the text.
func (r *Rope) Len() int { return r.root.length() }

// Index returns the byte at position i; ok is false outside 0..Len-1.
func (r *Rope) Index(i int) (byte, bool) {
	if i < 0 || i >= r.Len() {
		return 0, false
	}
	n := r.root
	for n.left != nil || n.right != nil {
		if i < n.weight {
			n = n.left
		} else {
			i -= n.weight
			n = n.right
		}
	}
	return n.text[i], true
}

func split(n *ropeNode, i int) (*ropeNode, *ropeNode) {
	if n == nil {
		return nil, nil
	}
	if n.left == nil && n.right == nil {
		return build(n.text[:i]), build(n.text[i:])
	}
	if i < n.weight {
		a, b := split(n.left, i)
		return a, join(b, n.right)
	}
	a, b := split(n.right, i-n.weight)
	return join(n.left, a), b
}

// Split cuts r at i (clamped to 0..Len) and returns the two halves; r is unchanged.
func (r *Rope) Split(i int) (*Rope, *Rope) {
	a, b := split(r.root, min(max(i, 0), r.Len()))
	return &Rope{a}, &Rope{b}
}

// Concat returns a rope of r followed by s, sharing both.
func (r *Rope) Concat(s *Rope) *Rope { return &Rope{join(r.root, s.root)} }

// Insert puts s at position i (clamped to 0..Len).
func (r *Rope) Insert(i int, s string) {
	a, b := r.Split(i)
	r.root = join(join(a.root, build(s)), b.root)
}

// Delete removes up to n bytes from position i and returns how many it removed.
func (r *Rope) Delete(i, n int) int {
	i = min(max(i, 0), r.Len())
	n = min(max(n, 0), r.Len()-i)
	a, rest := r.Split(i)
	_, b := rest.Split(n)
	r.root = join(a.root, b.root)
	return n
}

// String returns the text.
func (r *Rope) String() string {
	var b strings.Builder
	var walk func(*ropeNode)
	walk = func(n *ropeNode) {
		if n == nil {
			return
		}
		b.WriteString(n.text)
		walk(n.left)
		walk(n.right)
	}
	walk(r.root)
	return b.String()
}

// Depth returns the number of levels of the tree.
func (r *Rope) Depth() int {
	var depth func(*ropeNode) int
	depth = func(n *ropeNode) int {
		if n == nil {
			return 0
		}
		return 1 + max(depth(n.left), depth(n.right))
	}
	return depth(r.root)
}

// Rebalance rebuilds r as a balanced rope of the same text.
func (r *Rope) Rebalance() { r.root = build(r.String()) }

// Snapshot describes the tree: inner nodes show their weight, leaves their text.
func (r *Rope) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	var walk func(n *ropeNode, parent int, side string)
	walk = func(n *ropeNode, parent int, side string) {
		if n == nil {
			return
		}
		id := len(s.Nodes)
		label := fmt.Sprintf("weight %d", n.weight)
		if n.left == nil && n.right == nil {
			label = "\"" + n.text + "\""
		}
		nd := trace.Node{ID: id, Label: label, Parent: parent}
		if parent < 0 {
			nd.Role = "root"
		} else {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id, Label: side})
		}
		s.Nodes = append(s.Nodes, nd)
		walk(n.left, id, "left")
		walk(n.right, id, "right")
	}
	walk(r.root, -1, "")
	return s
}
