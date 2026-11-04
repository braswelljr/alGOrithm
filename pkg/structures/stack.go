package structures

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Stack is last in, first out: Push adds on top, Pop takes from the top. The
// zero Stack is empty and ready.
type Stack[T any] struct{ items []T }

// NewStack returns an empty stack holding values.
func NewStack[T any](values ...T) *Stack[T] { return &Stack[T]{items: append([]T(nil), values...)} }

// Push adds v on top. Time: O(1) amortized.
func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

// Pop removes and returns the top; ok is false, and nothing changes, when the
// stack is empty. Time: O(1).
func (s *Stack[T]) Pop() (v T, ok bool) {
	if len(s.items) == 0 {
		return v, false
	}
	v = s.items[len(s.items)-1]
	var zero T
	s.items[len(s.items)-1] = zero
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// Peek returns the top without removing it; ok is false when empty. Time: O(1).
func (s *Stack[T]) Peek() (v T, ok bool) {
	if len(s.items) == 0 {
		return v, false
	}
	return s.items[len(s.items)-1], true
}

// Len is the number of values. Time: O(1).
func (s *Stack[T]) Len() int { return len(s.items) }

// Values returns the values from bottom to top. Time: O(n).
func (s *Stack[T]) Values() []T { return append([]T(nil), s.items...) }

// Snapshot describes the stack bottom to top, the last node the "top".
func (s *Stack[T]) Snapshot() trace.Snapshot { return trace.Row(s.items, "", "", "top") }
