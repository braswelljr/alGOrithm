package persistent

import (
	"cmp"
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type treeNode[T cmp.Ordered] struct {
	value       T
	left, right *treeNode[T]
}

// Tree is an immutable binary search tree: Insert and Remove copy only the
// nodes on the path from the root to the change and point the copies at the
// untouched subtrees of the old version, so a new version costs O(h) nodes
// and every old version still works. The zero Tree is empty and ready to use.
//
// The tree is not rebalanced (that would copy more paths than it shows), so
// its height h is O(log n) for values in random order and O(n) for sorted ones.
//
// Time: O(h) Insert, Remove and Contains.
// Space: O(h) new nodes per update.
type Tree[T cmp.Ordered] struct {
	root *treeNode[T]
	n    int
}

// Insert returns t with v added (t itself when v is there already).
func (t Tree[T]) Insert(v T) Tree[T] {
	added := false
	var ins func(n *treeNode[T]) *treeNode[T]
	ins = func(n *treeNode[T]) *treeNode[T] {
		switch {
		case n == nil:
			added = true
			return &treeNode[T]{value: v}
		case v < n.value:
			if l := ins(n.left); added {
				return &treeNode[T]{value: n.value, left: l, right: n.right}
			}
		case v > n.value:
			if r := ins(n.right); added {
				return &treeNode[T]{value: n.value, left: n.left, right: r}
			}
		}
		return n
	}
	root := ins(t.root)
	if !added {
		return t
	}
	return Tree[T]{root: root, n: t.n + 1}
}

// Remove returns t without v (t itself when v is not there).
func (t Tree[T]) Remove(v T) Tree[T] {
	removed := false
	var del func(n *treeNode[T], v T) *treeNode[T]
	del = func(n *treeNode[T], v T) *treeNode[T] {
		switch {
		case n == nil:
			return nil
		case v < n.value:
			if l := del(n.left, v); removed {
				return &treeNode[T]{value: n.value, left: l, right: n.right}
			}
			return n
		case v > n.value:
			if r := del(n.right, v); removed {
				return &treeNode[T]{value: n.value, left: n.left, right: r}
			}
			return n
		}
		removed = true
		if n.left == nil {
			return n.right
		}
		if n.right == nil {
			return n.left
		}
		m := n.right
		for m.left != nil {
			m = m.left
		}
		return &treeNode[T]{value: m.value, left: n.left, right: del(n.right, m.value)}
	}
	root := del(t.root, v)
	if !removed {
		return t
	}
	return Tree[T]{root: root, n: t.n - 1}
}

// Contains reports whether v is in t.
func (t Tree[T]) Contains(v T) bool {
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

// Len returns the number of values.
func (t Tree[T]) Len() int { return t.n }

// Values returns the values in order.
func (t Tree[T]) Values() []T {
	out := make([]T, 0, t.n)
	var walk func(*treeNode[T])
	walk = func(n *treeNode[T]) {
		if n != nil {
			walk(n.left)
			out = append(out, n.value)
			walk(n.right)
		}
	}
	walk(t.root)
	return out
}

// SharedNodes returns how many of u's nodes are also nodes of t.
func (t Tree[T]) SharedNodes(u Tree[T]) int {
	in := map[*treeNode[T]]bool{}
	var mark func(*treeNode[T])
	mark = func(n *treeNode[T]) {
		if n != nil {
			in[n] = true
			mark(n.left)
			mark(n.right)
		}
	}
	mark(t.root)
	shared := 0
	var count func(*treeNode[T])
	count = func(n *treeNode[T]) {
		if n == nil {
			return
		}
		if in[n] {
			shared++
		}
		count(n.left)
		count(n.right)
	}
	count(u.root)
	return shared
}

// SnapshotTrees describes several versions at once, drawing each node once,
// so the subtrees they share show as shared.
func SnapshotTrees[T cmp.Ordered](vs ...Tree[T]) trace.Snapshot {
	names := make([]string, len(vs))
	for i := range vs {
		names[i] = versionName(i)
	}
	return versions(names, func(i int, visit visitFunc) {
		var walk func(n *treeNode[T], parent any, side string)
		walk = func(n *treeNode[T], parent any, side string) {
			if n == nil {
				return
			}
			if !visit(n, fmt.Sprint(n.value), parent, side) {
				return // shared with a version already drawn
			}
			walk(n.left, n, "left")
			walk(n.right, n, "right")
		}
		walk(vs[i].root, nil, "root")
	})
}
