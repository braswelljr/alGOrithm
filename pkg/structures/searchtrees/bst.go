package searchtrees

import (
	"cmp"
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type bstNode[T cmp.Ordered] struct {
	value       T
	left, right *bstNode[T]
}

// BST is a binary search tree of distinct values: smaller values to the left,
// larger to the right, so a search follows one path down. It does not
// rebalance; AVL does. The zero BST is empty and ready.
type BST[T cmp.Ordered] struct {
	root *bstNode[T]
	size int
}

// NewBST returns a tree holding values, inserted in order.
func NewBST[T cmp.Ordered](values ...T) *BST[T] {
	t := &BST[T]{}
	for _, v := range values {
		t.Insert(v)
	}
	return t
}

// Len is the number of values. Time: O(1).
func (t *BST[T]) Len() int { return t.size }

// Insert adds v and reports whether it was new. Time: O(h) for height h.
func (t *BST[T]) Insert(v T) bool {
	link := &t.root
	for *link != nil {
		switch n := *link; {
		case v < n.value:
			link = &n.left
		case v > n.value:
			link = &n.right
		default:
			return false
		}
	}
	*link = &bstNode[T]{value: v}
	t.size++
	return true
}

// Contains reports whether v is in the tree. Time: O(h).
func (t *BST[T]) Contains(v T) bool {
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

// Remove deletes v and reports whether it was there; a node with two children
// takes its in-order successor's value. Time: O(h).
func (t *BST[T]) Remove(v T) bool {
	var removed bool
	var del func(n *bstNode[T], v T) *bstNode[T]
	del = func(n *bstNode[T], v T) *bstNode[T] {
		switch {
		case n == nil:
			return nil
		case v < n.value:
			n.left = del(n.left, v)
		case v > n.value:
			n.right = del(n.right, v)
		default:
			removed = true
			if n.left == nil {
				return n.right
			}
			if n.right == nil {
				return n.left
			}
			s := n.right
			for s.left != nil {
				s = s.left
			}
			n.value = s.value
			n.right = del(n.right, s.value)
		}
		return n
	}
	t.root = del(t.root, v)
	if removed {
		t.size--
	}
	return removed
}

// Min returns the smallest value; ok is false when empty. Time: O(h).
func (t *BST[T]) Min() (v T, ok bool) {
	if t.root == nil {
		return v, false
	}
	n := t.root
	for n.left != nil {
		n = n.left
	}
	return n.value, true
}

// Max returns the largest value; ok is false when empty. Time: O(h).
func (t *BST[T]) Max() (v T, ok bool) {
	if t.root == nil {
		return v, false
	}
	n := t.root
	for n.right != nil {
		n = n.right
	}
	return n.value, true
}

// Height is the number of levels; 0 when empty. Time: O(n).
func (t *BST[T]) Height() int {
	var h func(*bstNode[T]) int
	h = func(n *bstNode[T]) int {
		if n == nil {
			return 0
		}
		return 1 + max(h(n.left), h(n.right))
	}
	return h(t.root)
}

// Values returns the values in order (an in-order walk). Time: O(n).
func (t *BST[T]) Values() []T {
	out := make([]T, 0, t.size)
	var walk func(*bstNode[T])
	walk = func(n *bstNode[T]) {
		if n != nil {
			walk(n.left)
			out = append(out, n.value)
			walk(n.right)
		}
	}
	walk(t.root)
	return out
}

// Snapshot describes the tree in pre-order, edges labeled "left" and "right".
func (t *BST[T]) Snapshot() trace.Snapshot {
	return binarySnapshot(t.root, func(n *bstNode[T]) (*bstNode[T], *bstNode[T]) { return n.left, n.right },
		func(n *bstNode[T], isRoot bool) (string, string) { return fmt.Sprint(n.value), rootRole(isRoot, "") })
}
