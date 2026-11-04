package lookup

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Set holds each value at most once, on a Go map. The zero Set is not ready;
// use NewSet.
type Set[T comparable] struct{ m map[T]struct{} }

// NewSet returns a set holding values.
func NewSet[T comparable](values ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{}, len(values))}
	for _, v := range values {
		s.m[v] = struct{}{}
	}
	return s
}

// Add puts v in the set and reports whether it was new. Time: O(1) average.
func (s *Set[T]) Add(v T) bool {
	if _, ok := s.m[v]; ok {
		return false
	}
	s.m[v] = struct{}{}
	return true
}

// Remove takes v out and reports whether it was there. Time: O(1) average.
func (s *Set[T]) Remove(v T) bool {
	_, ok := s.m[v]
	delete(s.m, v)
	return ok
}

// Contains reports whether v is in the set. Time: O(1) average.
func (s *Set[T]) Contains(v T) bool { _, ok := s.m[v]; return ok }

// Len is the number of values. Time: O(1).
func (s *Set[T]) Len() int { return len(s.m) }

// Values returns the values in no particular order. Time: O(n).
func (s *Set[T]) Values() []T {
	out := make([]T, 0, len(s.m))
	for v := range s.m {
		out = append(out, v)
	}
	return out
}

// Union returns a new set of the values in s or o. Time: O(n + m).
func (s *Set[T]) Union(o *Set[T]) *Set[T] {
	out := NewSet(s.Values()...)
	for v := range o.m {
		out.m[v] = struct{}{}
	}
	return out
}

// Intersection returns a new set of the values in both. Time: O(min(n, m)).
func (s *Set[T]) Intersection(o *Set[T]) *Set[T] {
	small, big := s, o
	if big.Len() < small.Len() {
		small, big = big, small
	}
	out := NewSet[T]()
	for v := range small.m {
		if big.Contains(v) {
			out.m[v] = struct{}{}
		}
	}
	return out
}

// Difference returns a new set of the values in s but not in o. Time: O(n).
func (s *Set[T]) Difference(o *Set[T]) *Set[T] {
	out := NewSet[T]()
	for v := range s.m {
		if !o.Contains(v) {
			out.m[v] = struct{}{}
		}
	}
	return out
}

// SubsetOf reports whether every value of s is in o. Time: O(n).
func (s *Set[T]) SubsetOf(o *Set[T]) bool {
	for v := range s.m {
		if !o.Contains(v) {
			return false
		}
	}
	return true
}

// Snapshot describes the members, sorted by their text so the picture does
// not depend on map order.
func (s *Set[T]) Snapshot() trace.Snapshot {
	labels := make([]string, 0, len(s.m))
	for v := range s.m {
		labels = append(labels, fmt.Sprint(v))
	}
	sort.Strings(labels)
	snap := trace.Snapshot{Layout: "set", Nodes: make([]trace.Node, len(labels))}
	for i, l := range labels {
		snap.Nodes[i] = trace.Node{ID: i, Label: l, Parent: -1}
	}
	return snap
}
