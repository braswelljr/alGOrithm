package structures

import (
	"cmp"
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MinStack is a LIFO stack that also answers "what is the smallest value on
// the stack?" in O(1): beside each value it keeps the minimum of everything
// at or below it. The zero MinStack is empty and ready.
type MinStack[T cmp.Ordered] struct {
	values, mins []T
}

// NewMinStack returns a stack holding values, the last on top.
func NewMinStack[T cmp.Ordered](values ...T) *MinStack[T] {
	s := &MinStack[T]{}
	for _, v := range values {
		s.Push(v)
	}
	return s
}

// Push adds v on top. Time: O(1) amortized.
func (s *MinStack[T]) Push(v T) {
	m := v
	if n := len(s.mins); n > 0 && s.mins[n-1] < m {
		m = s.mins[n-1]
	}
	s.values, s.mins = append(s.values, v), append(s.mins, m)
}

// Pop removes and returns the top; ok is false when empty. Time: O(1).
func (s *MinStack[T]) Pop() (v T, ok bool) {
	n := len(s.values)
	if n == 0 {
		return v, false
	}
	v = s.values[n-1]
	s.values, s.mins = s.values[:n-1], s.mins[:n-1]
	return v, true
}

// Peek returns the top without removing it; ok is false when empty.
func (s *MinStack[T]) Peek() (v T, ok bool) {
	if len(s.values) == 0 {
		return v, false
	}
	return s.values[len(s.values)-1], true
}

// Min returns the smallest value on the stack; ok is false when empty.
// Time: O(1).
func (s *MinStack[T]) Min() (v T, ok bool) {
	if len(s.mins) == 0 {
		return v, false
	}
	return s.mins[len(s.mins)-1], true
}

// Len is the number of values. Time: O(1).
func (s *MinStack[T]) Len() int { return len(s.values) }

// Snapshot describes the stack bottom to top, each value labeled with the
// minimum at that height ("5 (min 3)").
func (s *MinStack[T]) Snapshot() trace.Snapshot {
	labels := make([]string, len(s.values))
	for i, v := range s.values {
		labels[i] = fmt.Sprintf("%v (min %v)", v, s.mins[i])
	}
	return trace.Row(labels, "", "", "top")
}

var _ LIFO[int] = (*MinStack[int])(nil)
