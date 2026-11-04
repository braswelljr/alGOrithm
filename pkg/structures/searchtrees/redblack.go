package searchtrees

import (
	"cmp"
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type rbNode[T cmp.Ordered] struct {
	value       T
	left, right *rbNode[T]
	red         bool
}

// RedBlackTree is a self-balancing binary search tree of distinct values,
// written as a left-leaning red-black tree (Sedgewick, 2008): every node is
// red or black, a red node is always a left child, no red node has a red
// child, and every path from the root to an empty link passes the same number
// of black nodes. That keeps the height under 2 log n. The zero RedBlackTree
// is empty and ready.
type RedBlackTree[T cmp.Ordered] struct {
	root *rbNode[T]
	size int
}

// NewRedBlackTree returns a tree holding values.
func NewRedBlackTree[T cmp.Ordered](values ...T) *RedBlackTree[T] {
	t := &RedBlackTree[T]{}
	for _, v := range values {
		t.Insert(v)
	}
	return t
}

// Len is the number of values. Time: O(1).
func (t *RedBlackTree[T]) Len() int { return t.size }

func rbRed[T cmp.Ordered](n *rbNode[T]) bool { return n != nil && n.red }

func rbRotateLeft[T cmp.Ordered](h *rbNode[T]) *rbNode[T] {
	x := h.right
	h.right, x.left = x.left, h
	x.red, h.red = h.red, true
	return x
}

func rbRotateRight[T cmp.Ordered](h *rbNode[T]) *rbNode[T] {
	x := h.left
	h.left, x.right = x.right, h
	x.red, h.red = h.red, true
	return x
}

func rbFlip[T cmp.Ordered](h *rbNode[T]) {
	h.red = !h.red
	h.left.red = !h.left.red
	h.right.red = !h.right.red
}

// rbFixUp restores the left-leaning invariants on the way back up.
func rbFixUp[T cmp.Ordered](h *rbNode[T]) *rbNode[T] {
	if rbRed(h.right) && !rbRed(h.left) {
		h = rbRotateLeft(h)
	}
	if rbRed(h.left) && rbRed(h.left.left) {
		h = rbRotateRight(h)
	}
	if rbRed(h.left) && rbRed(h.right) {
		rbFlip(h)
	}
	return h
}

// Insert adds v and reports whether it was new. Time: O(log n).
func (t *RedBlackTree[T]) Insert(v T) bool {
	if t.Contains(v) {
		return false
	}
	var ins func(h *rbNode[T]) *rbNode[T]
	ins = func(h *rbNode[T]) *rbNode[T] {
		if h == nil {
			return &rbNode[T]{value: v, red: true}
		}
		if v < h.value {
			h.left = ins(h.left)
		} else {
			h.right = ins(h.right)
		}
		return rbFixUp(h)
	}
	t.root = ins(t.root)
	t.root.red = false
	t.size++
	return true
}

// Contains reports whether v is in the tree. Time: O(log n).
func (t *RedBlackTree[T]) Contains(v T) bool {
	for n := t.root; n != nil; {
		switch {
		case v < n.value:
			n = n.left
		case v > n.value:
			n = n.right
		default:
			return true
		}
	}
	return false
}

func rbMoveRedLeft[T cmp.Ordered](h *rbNode[T]) *rbNode[T] {
	rbFlip(h)
	if rbRed(h.right.left) {
		h.right = rbRotateRight(h.right)
		h = rbRotateLeft(h)
		rbFlip(h)
	}
	return h
}

func rbMoveRedRight[T cmp.Ordered](h *rbNode[T]) *rbNode[T] {
	rbFlip(h)
	if rbRed(h.left.left) {
		h = rbRotateRight(h)
		rbFlip(h)
	}
	return h
}

func rbDeleteMin[T cmp.Ordered](h *rbNode[T]) *rbNode[T] {
	if h.left == nil {
		return nil
	}
	if !rbRed(h.left) && !rbRed(h.left.left) {
		h = rbMoveRedLeft(h)
	}
	h.left = rbDeleteMin(h.left)
	return rbFixUp(h)
}

// Remove deletes v and reports whether it was there. Time: O(log n).
func (t *RedBlackTree[T]) Remove(v T) bool {
	if !t.Contains(v) {
		return false
	}
	var del func(h *rbNode[T]) *rbNode[T]
	del = func(h *rbNode[T]) *rbNode[T] {
		if v < h.value {
			if !rbRed(h.left) && !rbRed(h.left.left) {
				h = rbMoveRedLeft(h)
			}
			h.left = del(h.left)
			return rbFixUp(h)
		}
		if rbRed(h.left) {
			h = rbRotateRight(h)
		}
		if v == h.value && h.right == nil {
			return nil
		}
		if !rbRed(h.right) && !rbRed(h.right.left) {
			h = rbMoveRedRight(h)
		}
		if v == h.value {
			m := h.right
			for m.left != nil {
				m = m.left
			}
			h.value = m.value
			h.right = rbDeleteMin(h.right)
		} else {
			h.right = del(h.right)
		}
		return rbFixUp(h)
	}
	t.root = del(t.root)
	if t.root != nil {
		t.root.red = false
	}
	t.size--
	return true
}

// Values returns the values in ascending order. Time: O(n).
func (t *RedBlackTree[T]) Values() []T {
	out := make([]T, 0, t.size)
	var walk func(*rbNode[T])
	walk = func(n *rbNode[T]) {
		if n != nil {
			walk(n.left)
			out = append(out, n.value)
			walk(n.right)
		}
	}
	walk(t.root)
	return out
}

// Height is the number of levels; 0 when empty. Time: O(n).
func (t *RedBlackTree[T]) Height() int {
	var h func(*rbNode[T]) int
	h = func(n *rbNode[T]) int {
		if n == nil {
			return 0
		}
		return 1 + max(h(n.left), h(n.right))
	}
	return h(t.root)
}

// Valid reports whether the tree keeps every invariant: ordered, a black
// root, red links leaning left only, no two reds in a row, and the same
// number of black nodes on every path. Time: O(n).
func (t *RedBlackTree[T]) Valid() bool {
	if rbRed(t.root) {
		return false
	}
	var check func(n *rbNode[T], lo, hi *T) (blacks int, ok bool)
	check = func(n *rbNode[T], lo, hi *T) (int, bool) {
		if n == nil {
			return 1, true
		}
		if (lo != nil && n.value <= *lo) || (hi != nil && n.value >= *hi) {
			return 0, false
		}
		if rbRed(n.right) || (n.red && rbRed(n.left)) {
			return 0, false
		}
		l, okl := check(n.left, lo, &n.value)
		r, okr := check(n.right, &n.value, hi)
		if !okl || !okr || l != r {
			return 0, false
		}
		if !n.red {
			l++
		}
		return l, true
	}
	_, ok := check(t.root, nil, nil)
	return ok
}

// Snapshot describes the tree in pre-order, each node's Role its color.
func (t *RedBlackTree[T]) Snapshot() trace.Snapshot {
	return binarySnapshot(t.root, func(n *rbNode[T]) (*rbNode[T], *rbNode[T]) { return n.left, n.right },
		func(n *rbNode[T], _ bool) (string, string) {
			if n.red {
				return fmt.Sprint(n.value), "red"
			}
			return fmt.Sprint(n.value), "black"
		})
}
