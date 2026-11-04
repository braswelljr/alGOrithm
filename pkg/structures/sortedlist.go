package structures

import (
	"cmp"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// SortedList keeps its values in ascending order as they are inserted, with
// binary search to find where each belongs. Duplicates are kept. The zero
// SortedList is empty and ready.
type SortedList[T cmp.Ordered] struct{ items []T }

// NewSortedList returns a list holding values, sorted.
func NewSortedList[T cmp.Ordered](values ...T) *SortedList[T] {
	l := &SortedList[T]{}
	for _, v := range values {
		l.Insert(v)
	}
	return l
}

// Insert adds v in its place and returns its index. Time: O(log n) to find
// the place, O(n) to shift the values after it.
func (l *SortedList[T]) Insert(v T) int {
	i := sort.Search(len(l.items), func(i int) bool { return l.items[i] > v }) // after equal values: stable
	var zero T
	l.items = append(l.items, zero)
	copy(l.items[i+1:], l.items[i:])
	l.items[i] = v
	return i
}

// Index returns the index of v, or -1. Time: O(log n).
func (l *SortedList[T]) Index(v T) int {
	i := sort.Search(len(l.items), func(i int) bool { return l.items[i] >= v })
	if i < len(l.items) && l.items[i] == v {
		return i
	}
	return -1
}

// Remove deletes one v and reports whether there was one. Time: O(n).
func (l *SortedList[T]) Remove(v T) bool {
	i := l.Index(v)
	if i < 0 {
		return false
	}
	l.items = append(l.items[:i], l.items[i+1:]...)
	return true
}

// Len is the number of values. Time: O(1).
func (l *SortedList[T]) Len() int { return len(l.items) }

// Values returns the values in ascending order. Time: O(n).
func (l *SortedList[T]) Values() []T { return append([]T(nil), l.items...) }

// Snapshot describes the list smallest to largest.
func (l *SortedList[T]) Snapshot() trace.Snapshot { return trace.Row(l.items, "next", "min", "max") }
