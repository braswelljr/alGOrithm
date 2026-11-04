package searchtrees

import (
	"cmp"
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type splayNode[T cmp.Ordered] struct {
	value       T
	left, right *splayNode[T]
}

// SplayTree is a binary search tree that moves every value it touches to the
// root (Daniel Sleator and Robert Tarjan, 1985). It keeps no balance
// information at all, yet any sequence of m operations costs O(m log n), and
// values used often stay near the top, so it adapts to what is asked.
//
// Contains also splays, so even a search changes the tree's shape.
//
// Time: O(log n) amortized insert, remove and search; one operation can be O(n).
// Space: O(n).
type SplayTree[T cmp.Ordered] struct {
	root *splayNode[T]
	n    int
}

// NewSplayTree returns an empty splay tree.
func NewSplayTree[T cmp.Ordered]() *SplayTree[T] { return &SplayTree[T]{} }

// splay brings v, or the last node on the path to where v would be, to the
// root, with top-down splaying.
func (t *SplayTree[T]) splay(v T) {
	if t.root == nil {
		return
	}
	var header splayNode[T]
	l, r := &header, &header
	x := t.root
loop:
	for {
		switch {
		case v < x.value:
			if x.left == nil {
				break loop
			}
			if v < x.left.value { // zig-zig: rotate right
				y := x.left
				x.left, y.right = y.right, x
				x = y
				if x.left == nil {
					break loop
				}
			}
			r.left, r, x = x, x, x.left // link right
		case v > x.value:
			if x.right == nil {
				break loop
			}
			if v > x.right.value { // zig-zig: rotate left
				y := x.right
				x.right, y.left = y.left, x
				x = y
				if x.right == nil {
					break loop
				}
			}
			l.right, l, x = x, x, x.right // link left
		default:
			break loop
		}
	}
	l.right, r.left = x.left, x.right
	x.left, x.right = header.right, header.left
	t.root = x
}

// Insert adds v and reports whether it was not there already; either way v
// ends at the root.
func (t *SplayTree[T]) Insert(v T) bool {
	if t.root == nil {
		t.root, t.n = &splayNode[T]{value: v}, 1
		return true
	}
	t.splay(v)
	if t.root.value == v {
		return false
	}
	n := &splayNode[T]{value: v}
	if v < t.root.value {
		n.left, n.right, t.root.left = t.root.left, t.root, nil
	} else {
		n.right, n.left, t.root.right = t.root.right, t.root, nil
	}
	t.root = n
	t.n++
	return true
}

// Remove deletes v and reports whether it was there.
func (t *SplayTree[T]) Remove(v T) bool {
	if t.root == nil {
		return false
	}
	t.splay(v)
	if t.root.value != v {
		return false
	}
	if t.root.left == nil {
		t.root = t.root.right
	} else {
		right := t.root.right
		t.root = t.root.left
		t.splay(v) // the largest value left of v comes up, with no right child
		t.root.right = right
	}
	t.n--
	return true
}

// Contains reports whether v is in the tree, splaying it (or its neighbor)
// to the root.
func (t *SplayTree[T]) Contains(v T) bool {
	t.splay(v)
	return t.root != nil && t.root.value == v
}

// Root returns the value at the root, the one touched last.
func (t *SplayTree[T]) Root() (T, bool) {
	if t.root == nil {
		var zero T
		return zero, false
	}
	return t.root.value, true
}

// Len returns the number of values.
func (t *SplayTree[T]) Len() int { return t.n }

// Values returns the values in order.
func (t *SplayTree[T]) Values() []T {
	out := make([]T, 0, t.n)
	var walk func(*splayNode[T])
	walk = func(n *splayNode[T]) {
		if n != nil {
			walk(n.left)
			out = append(out, n.value)
			walk(n.right)
		}
	}
	walk(t.root)
	return out
}

// Snapshot describes the tree, root first.
func (t *SplayTree[T]) Snapshot() trace.Snapshot {
	return binarySnapshot(t.root, func(n *splayNode[T]) (*splayNode[T], *splayNode[T]) { return n.left, n.right },
		func(n *splayNode[T], isRoot bool) (string, string) { return fmt.Sprint(n.value), rootRole(isRoot, "") })
}
