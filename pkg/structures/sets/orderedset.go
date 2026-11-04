package sets

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// OrderedSet is a set kept as a sorted slice of unique values: membership,
// rank, floor and ceiling by binary search, and ranges in order. The zero
// OrderedSet is empty and ready.
type OrderedSet[T cmp.Ordered] struct{ items []T }

// NewOrderedSet returns a set holding values.
func NewOrderedSet[T cmp.Ordered](values ...T) *OrderedSet[T] {
	s := &OrderedSet[T]{}
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Add inserts v and reports whether it was new. Time: O(log n) to find the
// place, O(n) to shift.
func (s *OrderedSet[T]) Add(v T) bool {
	i, found := slices.BinarySearch(s.items, v)
	if found {
		return false
	}
	s.items = slices.Insert(s.items, i, v)
	return true
}

// Remove deletes v and reports whether it was there. Time: O(n).
func (s *OrderedSet[T]) Remove(v T) bool {
	i, found := slices.BinarySearch(s.items, v)
	if found {
		s.items = slices.Delete(s.items, i, i+1)
	}
	return found
}

// Contains reports whether v is in the set. Time: O(log n).
func (s *OrderedSet[T]) Contains(v T) bool {
	_, found := slices.BinarySearch(s.items, v)
	return found
}

// Rank is the number of members smaller than v. Time: O(log n).
func (s *OrderedSet[T]) Rank(v T) int {
	i, _ := slices.BinarySearch(s.items, v)
	return i
}

// Floor returns the largest member <= v; ok is false when there is none.
func (s *OrderedSet[T]) Floor(v T) (T, bool) {
	i, found := slices.BinarySearch(s.items, v)
	if found {
		return s.items[i], true
	}
	if i == 0 {
		var zero T
		return zero, false
	}
	return s.items[i-1], true
}

// Ceiling returns the smallest member >= v; ok is false when there is none.
func (s *OrderedSet[T]) Ceiling(v T) (T, bool) {
	i, _ := slices.BinarySearch(s.items, v)
	if i == len(s.items) {
		var zero T
		return zero, false
	}
	return s.items[i], true
}

// Range returns the members in [lo, hi], in order. Time: O(log n + k).
func (s *OrderedSet[T]) Range(lo, hi T) []T {
	i, _ := slices.BinarySearch(s.items, lo)
	var out []T
	for ; i < len(s.items) && s.items[i] <= hi; i++ {
		out = append(out, s.items[i])
	}
	return out
}

// Len is the number of members.
func (s *OrderedSet[T]) Len() int { return len(s.items) }

// Values returns the members in ascending order.
func (s *OrderedSet[T]) Values() []T { return slices.Clone(s.items) }

// Snapshot describes the members in order, the ends marked "min" and "max".
func (s *OrderedSet[T]) Snapshot() trace.Snapshot {
	snap := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, len(s.items))}
	for i, v := range s.items {
		snap.Nodes[i] = trace.Node{ID: i, Label: fmt.Sprint(v), Parent: -1}
	}
	if n := len(snap.Nodes); n > 0 {
		snap.Nodes[0].Role = "min"
		if n > 1 {
			snap.Nodes[n-1].Role = "max"
		}
	}
	return snap
}
