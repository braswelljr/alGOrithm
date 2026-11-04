package searchtrees

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type bNode[T cmp.Ordered] struct {
	keys     []T
	children []*bNode[T]
}

func (n *bNode[T]) leaf() bool { return len(n.children) == 0 }

// BTree is a B-tree of minimum degree t (Bayer and McCreight, 1970): every
// node but the root holds t-1 to 2t-1 sorted keys, an inner node with k keys
// has k+1 children, and all leaves are at the same depth. Wide nodes keep the
// tree shallow, which is why databases and file systems use it. Values are
// distinct. The zero BTree is not ready; use NewBTree.
type BTree[T cmp.Ordered] struct {
	t    int
	root *bNode[T]
	size int
}

// NewBTree returns an empty tree of minimum degree t (at least 2) holding
// values.
func NewBTree[T cmp.Ordered](t int, values ...T) *BTree[T] {
	b := &BTree[T]{t: max(2, t), root: &bNode[T]{}}
	for _, v := range values {
		b.Insert(v)
	}
	return b
}

// Len is the number of keys; Degree the minimum degree. Time: O(1).
func (b *BTree[T]) Len() int { return b.size }

// Degree is the tree's minimum degree t.
func (b *BTree[T]) Degree() int { return b.t }

// Contains reports whether v is in the tree. Time: O(t log_t n).
func (b *BTree[T]) Contains(v T) bool {
	for n := b.root; ; {
		i, found := slices.BinarySearch(n.keys, v)
		if found {
			return true
		}
		if n.leaf() {
			return false
		}
		n = n.children[i]
	}
}

// split splits x's full child i around its middle key, which moves up into x.
func (b *BTree[T]) split(x *bNode[T], i int) {
	t := b.t
	y := x.children[i]
	z := &bNode[T]{keys: slices.Clone(y.keys[t:])}
	if !y.leaf() {
		z.children = slices.Clone(y.children[t:])
		y.children = y.children[:t]
	}
	mid := y.keys[t-1]
	y.keys = y.keys[:t-1]
	x.children = slices.Insert(x.children, i+1, z)
	x.keys = slices.Insert(x.keys, i, mid)
}

// Insert adds v and reports whether it was new, splitting full nodes on the
// way down so there is always room. Time: O(t log_t n).
func (b *BTree[T]) Insert(v T) bool {
	if b.Contains(v) {
		return false
	}
	if len(b.root.keys) == 2*b.t-1 {
		b.root = &bNode[T]{children: []*bNode[T]{b.root}}
		b.split(b.root, 0)
	}
	n := b.root
	for {
		i, _ := slices.BinarySearch(n.keys, v)
		if n.leaf() {
			n.keys = slices.Insert(n.keys, i, v)
			break
		}
		if len(n.children[i].keys) == 2*b.t-1 {
			b.split(n, i)
			if v > n.keys[i] {
				i++
			}
		}
		n = n.children[i]
	}
	b.size++
	return true
}

// Remove deletes v and reports whether it was there. Each node it descends
// into is first given at least t keys (borrowing from a sibling or merging),
// so the delete never has to back up. Time: O(t log_t n).
func (b *BTree[T]) Remove(v T) bool {
	if !b.Contains(v) {
		return false
	}
	b.remove(b.root, v)
	if len(b.root.keys) == 0 && !b.root.leaf() {
		b.root = b.root.children[0]
	}
	b.size--
	return true
}

func (b *BTree[T]) remove(x *bNode[T], v T) {
	t := b.t
	i, found := slices.BinarySearch(x.keys, v)
	if found {
		switch {
		case x.leaf():
			x.keys = slices.Delete(x.keys, i, i+1)
		case len(x.children[i].keys) >= t:
			p := x.children[i]
			for !p.leaf() {
				p = p.children[len(p.children)-1]
			}
			pred := p.keys[len(p.keys)-1]
			x.keys[i] = pred
			b.remove(x.children[i], pred)
		case len(x.children[i+1].keys) >= t:
			s := x.children[i+1]
			for !s.leaf() {
				s = s.children[0]
			}
			succ := s.keys[0]
			x.keys[i] = succ
			b.remove(x.children[i+1], succ)
		default:
			b.merge(x, i)
			b.remove(x.children[i], v)
		}
		return
	}
	if x.leaf() {
		return
	}
	if len(x.children[i].keys) == t-1 {
		switch {
		case i > 0 && len(x.children[i-1].keys) >= t:
			c, s := x.children[i], x.children[i-1]
			c.keys = slices.Insert(c.keys, 0, x.keys[i-1])
			x.keys[i-1] = s.keys[len(s.keys)-1]
			s.keys = s.keys[:len(s.keys)-1]
			if !s.leaf() {
				c.children = slices.Insert(c.children, 0, s.children[len(s.children)-1])
				s.children = s.children[:len(s.children)-1]
			}
		case i < len(x.children)-1 && len(x.children[i+1].keys) >= t:
			c, s := x.children[i], x.children[i+1]
			c.keys = append(c.keys, x.keys[i])
			x.keys[i] = s.keys[0]
			s.keys = slices.Delete(s.keys, 0, 1)
			if !s.leaf() {
				c.children = append(c.children, s.children[0])
				s.children = slices.Delete(s.children, 0, 1)
			}
		case i < len(x.children)-1:
			b.merge(x, i)
		default:
			b.merge(x, i-1)
			i--
		}
	}
	b.remove(x.children[i], v)
}

// merge joins x's children i and i+1 around x's key i.
func (b *BTree[T]) merge(x *bNode[T], i int) {
	l, r := x.children[i], x.children[i+1]
	l.keys = append(append(l.keys, x.keys[i]), r.keys...)
	l.children = append(l.children, r.children...)
	x.keys = slices.Delete(x.keys, i, i+1)
	x.children = slices.Delete(x.children, i+1, i+2)
}

// Values returns the keys in ascending order. Time: O(n).
func (b *BTree[T]) Values() []T {
	out := make([]T, 0, b.size)
	var walk func(*bNode[T])
	walk = func(n *bNode[T]) {
		for i, k := range n.keys {
			if !n.leaf() {
				walk(n.children[i])
			}
			out = append(out, k)
		}
		if !n.leaf() {
			walk(n.children[len(n.children)-1])
		}
	}
	walk(b.root)
	return out
}

// Height is the number of levels. Time: O(log_t n).
func (b *BTree[T]) Height() int {
	h := 1
	for n := b.root; !n.leaf(); n = n.children[0] {
		h++
	}
	return h
}

// Valid reports whether every node holds t-1 to 2t-1 keys (the root at most
// 2t-1), keys are sorted and between their parent's, inner nodes have one
// more child than keys, and all leaves are at the same depth. Time: O(n).
func (b *BTree[T]) Valid() bool {
	leafDepth := -1
	var check func(n *bNode[T], depth int, lo, hi *T) bool
	check = func(n *bNode[T], depth int, lo, hi *T) bool {
		if len(n.keys) > 2*b.t-1 || (n != b.root && len(n.keys) < b.t-1) {
			return false
		}
		for i, k := range n.keys {
			if (i > 0 && n.keys[i-1] >= k) || (lo != nil && k <= *lo) || (hi != nil && k >= *hi) {
				return false
			}
		}
		if n.leaf() {
			if leafDepth < 0 {
				leafDepth = depth
			}
			return depth == leafDepth
		}
		if len(n.children) != len(n.keys)+1 {
			return false
		}
		for i, c := range n.children {
			clo, chi := lo, hi
			if i > 0 {
				clo = &n.keys[i-1]
			}
			if i < len(n.keys) {
				chi = &n.keys[i]
			}
			if !check(c, depth+1, clo, chi) {
				return false
			}
		}
		return true
	}
	return check(b.root, 0, nil, nil)
}

// Snapshot describes the tree in pre-order, each node labeled with its keys.
func (b *BTree[T]) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	if len(b.root.keys) == 0 {
		return s
	}
	var walk func(n *bNode[T], parent int)
	walk = func(n *bNode[T], parent int) {
		id := len(s.Nodes)
		parts := make([]string, len(n.keys))
		for i, k := range n.keys {
			parts[i] = fmt.Sprint(k)
		}
		nd := trace.Node{ID: id, Label: strings.Join(parts, " "), Parent: parent}
		if parent < 0 {
			nd.Role = "root"
		} else {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id})
		}
		s.Nodes = append(s.Nodes, nd)
		for _, c := range n.children {
			walk(c, id)
		}
	}
	walk(b.root, -1)
	return s
}
