package persistent

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type stackNode[T any] struct {
	value T
	next  *stackNode[T]
	n     int // the length from here down
}

// Stack is an immutable LIFO: Push and Pop return a new Stack and never
// change the one they were called on. Every version that pushed onto s
// shares all of s, so n versions of a stack of length L take O(n + L)
// memory, not O(n * L). The zero Stack is empty and ready to use.
//
// Time: O(1) Push, Pop, Peek and Len.
// Space: O(1) per Push.
type Stack[T any] struct{ top *stackNode[T] }

// Push returns s with v on top; s itself is unchanged.
func (s Stack[T]) Push(v T) Stack[T] {
	return Stack[T]{&stackNode[T]{value: v, next: s.top, n: s.Len() + 1}}
}

// Pop returns the top value and s without it; ok is false when s is empty.
func (s Stack[T]) Pop() (v T, rest Stack[T], ok bool) {
	if s.top == nil {
		return v, s, false
	}
	return s.top.value, Stack[T]{s.top.next}, true
}

// Peek returns the top value; ok is false when s is empty.
func (s Stack[T]) Peek() (T, bool) {
	if s.top == nil {
		var zero T
		return zero, false
	}
	return s.top.value, true
}

// Len returns the number of values.
func (s Stack[T]) Len() int {
	if s.top == nil {
		return 0
	}
	return s.top.n
}

// Values returns the values from top to bottom.
func (s Stack[T]) Values() []T {
	out := make([]T, 0, s.Len())
	for n := s.top; n != nil; n = n.next {
		out = append(out, n.value)
	}
	return out
}

// Shares reports whether s and t share their bottom node, which they do when
// one was built from the other or both from a common version.
func (s Stack[T]) Shares(t Stack[T]) bool {
	bottom := func(n *stackNode[T]) *stackNode[T] {
		for n != nil && n.next != nil {
			n = n.next
		}
		return n
	}
	return s.top != nil && bottom(s.top) == bottom(t.top)
}

// Snapshot describes s alone, top first.
func (s Stack[T]) Snapshot() trace.Snapshot { return trace.Row(s.Values(), "next", "top", "bottom") }

// SnapshotStacks describes several versions at once, drawing each node
// once, so the nodes they share show as shared.
func SnapshotStacks[T any](vs ...Stack[T]) trace.Snapshot {
	names := make([]string, len(vs))
	for i := range vs {
		names[i] = versionName(i)
	}
	return versions(names, func(i int, visit visitFunc) {
		var parent any
		side := "top"
		for n := vs[i].top; n != nil; n = n.next {
			if !visit(n, fmt.Sprint(n.value), parent, side) {
				return // the rest is shared with a version already drawn
			}
			parent, side = n, "next"
		}
	})
}
