package sets

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Multiset is a bag: a set in which a value can occur more than once, stored
// as a count per value. The zero Multiset is not ready; use NewMultiset.
type Multiset[T cmp.Ordered] struct {
	counts map[T]int
	size   int
}

// NewMultiset returns a bag holding values, repeats counted.
func NewMultiset[T cmp.Ordered](values ...T) *Multiset[T] {
	m := &Multiset[T]{counts: map[T]int{}}
	for _, v := range values {
		m.Add(v)
	}
	return m
}

// Add puts one more v in the bag and returns its new count. Time: O(1) average.
func (m *Multiset[T]) Add(v T) int {
	m.counts[v]++
	m.size++
	return m.counts[v]
}

// Remove takes one v out and reports whether there was one. Time: O(1) average.
func (m *Multiset[T]) Remove(v T) bool {
	c := m.counts[v]
	if c == 0 {
		return false
	}
	if c == 1 {
		delete(m.counts, v)
	} else {
		m.counts[v] = c - 1
	}
	m.size--
	return true
}

// Count is how many times v is in the bag. Time: O(1) average.
func (m *Multiset[T]) Count(v T) int { return m.counts[v] }

// Len is the number of values counting repeats; Distinct without.
func (m *Multiset[T]) Len() int { return m.size }

// Distinct is the number of different values.
func (m *Multiset[T]) Distinct() int { return len(m.counts) }

// Values returns the distinct values in ascending order. Time: O(k log k).
func (m *Multiset[T]) Values() []T {
	out := make([]T, 0, len(m.counts))
	for v := range m.counts {
		out = append(out, v)
	}
	slices.Sort(out)
	return out
}

// Union returns a new bag with each value's larger count of the two (the
// multiset union). Time: O(k + l).
func (m *Multiset[T]) Union(o *Multiset[T]) *Multiset[T] {
	out := NewMultiset[T]()
	for _, src := range []*Multiset[T]{m, o} {
		for v, c := range src.counts {
			if c > out.counts[v] {
				out.size += c - out.counts[v]
				out.counts[v] = c
			}
		}
	}
	return out
}

// Sum returns a new bag with each value's counts added (the multiset sum).
// Time: O(k + l).
func (m *Multiset[T]) Sum(o *Multiset[T]) *Multiset[T] {
	out := NewMultiset[T]()
	for _, src := range []*Multiset[T]{m, o} {
		for v, c := range src.counts {
			out.counts[v] += c
			out.size += c
		}
	}
	return out
}

// Intersection returns a new bag with each value's smaller count.
func (m *Multiset[T]) Intersection(o *Multiset[T]) *Multiset[T] {
	out := NewMultiset[T]()
	for v, c := range m.counts {
		if k := min(c, o.counts[v]); k > 0 {
			out.counts[v], out.size = k, out.size+k
		}
	}
	return out
}

// Snapshot describes the distinct values in order as "value x count".
func (m *Multiset[T]) Snapshot() trace.Snapshot {
	vals := m.Values()
	s := trace.Snapshot{Layout: "set", Nodes: make([]trace.Node, len(vals))}
	for i, v := range vals {
		s.Nodes[i] = trace.Node{ID: i, Label: fmt.Sprintf("%v x%d", v, m.counts[v]), Parent: -1}
	}
	return s
}
