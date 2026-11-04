package databases

import (
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// bpNode is a node of a B+ tree. A leaf holds keys and their values and
// links to the next leaf; an inner node holds separator keys and one more
// child than keys, child i holding the keys below keys[i].
type bpNode struct {
	leaf     bool
	keys     []int
	vals     []int
	children []*bpNode
	next     *bpNode
}

// BPlusTree is a B+ tree index of order m (Bayer and McCreight's B-tree, with
// every value kept in the leaves and the leaves linked in key order): the
// index most databases put on a table's columns. Inner nodes only guide the
// search, so they hold many keys and the tree stays three or four levels deep
// for millions of rows; a range scan finds its first key once and then walks
// the linked leaves.
//
// Every node holds at most m-1 keys, and every node but the root at least
// half that; all leaves are at the same depth. Keys are distinct; inserting a
// key again replaces its value. Simplified: nodes live in memory, not in disk
// pages. The zero BPlusTree is not ready; use NewBPlusTree.
type BPlusTree struct {
	order int
	root  *bpNode
	size  int
}

// NewBPlusTree returns an empty tree of order m (at least 3, the most
// children an inner node has).
func NewBPlusTree(m int) *BPlusTree {
	return &BPlusTree{order: max(3, m), root: &bpNode{leaf: true}}
}

// Len is the number of keys. Time: O(1).
func (t *BPlusTree) Len() int { return t.size }

// Order is the most children an inner node has.
func (t *BPlusTree) Order() int { return t.order }

// Height is the number of levels, 1 for a tree that is one leaf.
func (t *BPlusTree) Height() int {
	h := 1
	for n := t.root; !n.leaf; n = n.children[0] {
		h++
	}
	return h
}

func (t *BPlusTree) minLeaf() int  { return t.order / 2 }
func (t *BPlusTree) minInner() int { return (t.order+1)/2 - 1 }

// childIndex is the child of an inner node to follow for key: the number of
// separators at most key.
func childIndex(n *bpNode, key int) int {
	i, found := slices.BinarySearch(n.keys, key)
	if found {
		i++
	}
	return i
}

// Search returns the value stored under key, and whether there is one,
// walking from the root to one leaf and reporting each node's keys to r as a
// Visit. Time: O(log_m n) nodes, O(log n) comparisons.
func (t *BPlusTree) Search(key int, r trace.Recorder) (int, bool) {
	n := t.root
	for {
		r.Record(trace.Event{Kind: trace.Visit, Values: slices.Clone(n.keys)})
		if n.leaf {
			i, found := slices.BinarySearch(n.keys, key)
			if !found {
				return 0, false
			}
			return n.vals[i], true
		}
		n = n.children[childIndex(n, key)]
	}
}

// Range returns the keys from lo to hi inclusive, in order: one walk down to
// the first leaf that could hold lo, then along the linked leaves. Time:
// O(log_m n + k) for k keys returned.
func (t *BPlusTree) Range(lo, hi int, r trace.Recorder) []int {
	n := t.root
	for !n.leaf {
		r.Record(trace.Event{Kind: trace.Visit, Values: slices.Clone(n.keys)})
		n = n.children[childIndex(n, lo)]
	}
	var out []int
	for ; n != nil; n = n.next {
		r.Record(trace.Event{Kind: trace.Visit, Values: slices.Clone(n.keys)})
		for _, k := range n.keys {
			if k > hi {
				return out
			}
			if k >= lo {
				out = append(out, k)
			}
		}
	}
	return out
}

// Keys returns every key in order, read along the leaves. Time: O(n).
func (t *BPlusTree) Keys() []int {
	n := t.root
	for !n.leaf {
		n = n.children[0]
	}
	var out []int
	for ; n != nil; n = n.next {
		out = append(out, n.keys...)
	}
	return out
}

// Insert stores value under key and reports whether the key is new. A full
// leaf splits in two and its first key moves up as a separator; a full inner
// node splits around its middle key; a split root makes the tree one level
// taller. Time: O(m log_m n).
func (t *BPlusTree) Insert(key, value int) bool {
	sep, right, added := t.insert(t.root, key, value)
	if right != nil {
		t.root = &bpNode{keys: []int{sep}, children: []*bpNode{t.root, right}}
	}
	if added {
		t.size++
	}
	return added
}

func (t *BPlusTree) insert(n *bpNode, key, value int) (sep int, right *bpNode, added bool) {
	if n.leaf {
		i, found := slices.BinarySearch(n.keys, key)
		if found {
			n.vals[i] = value
			return 0, nil, false
		}
		n.keys = slices.Insert(n.keys, i, key)
		n.vals = slices.Insert(n.vals, i, value)
		if len(n.keys) < t.order {
			return 0, nil, true
		}
		mid := len(n.keys) / 2
		r := &bpNode{leaf: true, keys: slices.Clone(n.keys[mid:]), vals: slices.Clone(n.vals[mid:]), next: n.next}
		n.keys, n.vals, n.next = n.keys[:mid:mid], n.vals[:mid:mid], r
		return r.keys[0], r, true
	}
	i := childIndex(n, key)
	s, r, added := t.insert(n.children[i], key, value)
	if r == nil {
		return 0, nil, added
	}
	n.keys = slices.Insert(n.keys, i, s)
	n.children = slices.Insert(n.children, i+1, r)
	if len(n.keys) < t.order {
		return 0, nil, added
	}
	mid := len(n.keys) / 2
	up := n.keys[mid]
	nr := &bpNode{keys: slices.Clone(n.keys[mid+1:]), children: slices.Clone(n.children[mid+1:])}
	n.keys, n.children = n.keys[:mid:mid], n.children[:mid+1:mid+1]
	return up, nr, added
}

// Delete removes key and reports whether it was there. A node left with too
// few keys borrows one from a sibling that can spare it, or else merges with
// a sibling, taking the separator down from the parent; a root left with one
// child is replaced by it, making the tree one level shorter. Time:
// O(m log_m n).
func (t *BPlusTree) Delete(key int) bool {
	if !t.delete(t.root, key) {
		return false
	}
	if !t.root.leaf && len(t.root.keys) == 0 {
		t.root = t.root.children[0]
	}
	t.size--
	return true
}

func (t *BPlusTree) delete(n *bpNode, key int) bool {
	if n.leaf {
		i, found := slices.BinarySearch(n.keys, key)
		if !found {
			return false
		}
		n.keys = slices.Delete(n.keys, i, i+1)
		n.vals = slices.Delete(n.vals, i, i+1)
		return true
	}
	i := childIndex(n, key)
	if !t.delete(n.children[i], key) {
		return false
	}
	t.rebalance(n, i)
	return true
}

func (t *BPlusTree) underflows(n *bpNode) bool {
	if n.leaf {
		return len(n.keys) < t.minLeaf()
	}
	return len(n.keys) < t.minInner()
}

func (t *BPlusTree) spare(n *bpNode) bool {
	if n.leaf {
		return len(n.keys) > t.minLeaf()
	}
	return len(n.keys) > t.minInner()
}

// rebalance fixes child i of p after a deletion left it too small.
func (t *BPlusTree) rebalance(p *bpNode, i int) {
	c := p.children[i]
	if !t.underflows(c) {
		return
	}
	switch {
	case i > 0 && t.spare(p.children[i-1]):
		l := p.children[i-1]
		last := len(l.keys) - 1
		if c.leaf {
			c.keys = slices.Insert(c.keys, 0, l.keys[last])
			c.vals = slices.Insert(c.vals, 0, l.vals[last])
			l.keys, l.vals = l.keys[:last], l.vals[:last]
			p.keys[i-1] = c.keys[0]
			return
		}
		c.keys = slices.Insert(c.keys, 0, p.keys[i-1])
		c.children = slices.Insert(c.children, 0, l.children[last+1])
		p.keys[i-1] = l.keys[last]
		l.keys, l.children = l.keys[:last], l.children[:last+1]
	case i+1 < len(p.children) && t.spare(p.children[i+1]):
		r := p.children[i+1]
		if c.leaf {
			c.keys = append(c.keys, r.keys[0])
			c.vals = append(c.vals, r.vals[0])
			r.keys, r.vals = slices.Delete(r.keys, 0, 1), slices.Delete(r.vals, 0, 1)
			p.keys[i] = r.keys[0]
			return
		}
		c.keys = append(c.keys, p.keys[i])
		c.children = append(c.children, r.children[0])
		p.keys[i] = r.keys[0]
		r.keys, r.children = slices.Delete(r.keys, 0, 1), slices.Delete(r.children, 0, 1)
	case i > 0:
		merge(p, i-1)
	default:
		merge(p, i)
	}
}

// merge joins children s and s+1 of p into child s, with separator s.
func merge(p *bpNode, s int) {
	l, r := p.children[s], p.children[s+1]
	if l.leaf {
		l.keys = append(l.keys, r.keys...)
		l.vals = append(l.vals, r.vals...)
		l.next = r.next
	} else {
		l.keys = append(append(l.keys, p.keys[s]), r.keys...)
		l.children = append(l.children, r.children...)
	}
	p.keys = slices.Delete(p.keys, s, s+1)
	p.children = slices.Delete(p.children, s+1, s+2)
}

// Snapshot draws the tree: each node's keys, inner nodes above their
// children, and "next" edges along the leaves.
func (t *BPlusTree) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	if t.size == 0 {
		return s
	}
	ids := map[*bpNode]int{}
	var leaves []*bpNode
	var walk func(n *bpNode, parent int)
	walk = func(n *bpNode, parent int) {
		id := len(s.Nodes)
		ids[n] = id
		parts := make([]string, len(n.keys))
		for i, k := range n.keys {
			parts[i] = strconv.Itoa(k)
		}
		nd := trace.Node{ID: id, Label: strings.Join(parts, " "), Parent: parent}
		switch {
		case parent < 0:
			nd.Role = "root"
		case n.leaf:
			nd.Role = "leaf"
		}
		if parent >= 0 {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id})
		}
		s.Nodes = append(s.Nodes, nd)
		if n.leaf {
			leaves = append(leaves, n)
		}
		for _, c := range n.children {
			walk(c, id)
		}
	}
	walk(t.root, -1)
	for i := 0; i+1 < len(leaves); i++ {
		s.Edges = append(s.Edges, trace.Edge{From: ids[leaves[i]], To: ids[leaves[i+1]], Label: "next"})
	}
	return s
}
