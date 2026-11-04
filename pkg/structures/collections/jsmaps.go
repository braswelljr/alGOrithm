package collections

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// JSMap is a JavaScript Map: keys of any comparable type, iterated in the order
// they were first set (a Go map has no order). Set on an existing key keeps
// its place; Delete and set again moves it to the end. The zero JSMap is empty
// and ready.
type JSMap[K comparable, V any] struct {
	index map[K]int // position in keys
	keys  []K
	vals  []V
	live  []bool
	size  int
}

// NewJSMap returns an empty JSMap.
func NewJSMap[K comparable, V any]() *JSMap[K, V] { return &JSMap[K, V]{} }

// Set adds or updates key (Map.prototype.set). Time: O(1) amortized.
func (m *JSMap[K, V]) Set(k K, v V) *JSMap[K, V] {
	if m.index == nil {
		m.index = map[K]int{}
	}
	if i, ok := m.index[k]; ok {
		m.vals[i] = v
		return m
	}
	m.index[k] = len(m.keys)
	m.keys, m.vals, m.live = append(m.keys, k), append(m.vals, v), append(m.live, true)
	m.size++
	return m
}

// Get returns key's value (Map.prototype.get); ok is false when absent.
func (m *JSMap[K, V]) Get(k K) (v V, ok bool) {
	i, ok := m.index[k]
	if !ok {
		return v, false
	}
	return m.vals[i], true
}

// Has reports whether key is present (Map.prototype.has).
func (m *JSMap[K, V]) Has(k K) bool { _, ok := m.index[k]; return ok }

// Delete removes key and reports whether it was there (Map.prototype.delete).
// Time: O(1); the slot is compacted away once half the slots are dead.
func (m *JSMap[K, V]) Delete(k K) bool {
	i, ok := m.index[k]
	if !ok {
		return false
	}
	delete(m.index, k)
	m.live[i] = false
	var zero V
	m.vals[i] = zero
	m.size--
	if len(m.keys) > 8 && m.size < len(m.keys)/2 {
		m.compact()
	}
	return true
}

func (m *JSMap[K, V]) compact() {
	keys, vals := m.Keys(), m.Values()
	*m = JSMap[K, V]{}
	for i := range keys {
		m.Set(keys[i], vals[i])
	}
}

// Clear removes everything (Map.prototype.clear).
func (m *JSMap[K, V]) Clear() { *m = JSMap[K, V]{} }

// Size is the number of keys (Map.prototype.size).
func (m *JSMap[K, V]) Size() int { return m.size }

// Keys returns the keys in insertion order (Map.prototype.keys).
func (m *JSMap[K, V]) Keys() []K {
	out := make([]K, 0, m.size)
	for i, k := range m.keys {
		if m.live[i] {
			out = append(out, k)
		}
	}
	return out
}

// Values returns the values in insertion order (Map.prototype.values).
func (m *JSMap[K, V]) Values() []V {
	out := make([]V, 0, m.size)
	for i, v := range m.vals {
		if m.live[i] {
			out = append(out, v)
		}
	}
	return out
}

// Entries returns the pairs in insertion order (Map.prototype.entries).
func (m *JSMap[K, V]) Entries() []Pair[K, V] {
	out := make([]Pair[K, V], 0, m.size)
	for i := range m.keys {
		if m.live[i] {
			out = append(out, Pair[K, V]{m.keys[i], m.vals[i]})
		}
	}
	return out
}

// ForEach calls f with each value and key in insertion order
// (Map.prototype.forEach).
func (m *JSMap[K, V]) ForEach(f func(v V, k K)) {
	for _, p := range m.Entries() {
		f(p.Value, p.Key)
	}
}

// Snapshot shows the entries in insertion order as "key => value".
func (m *JSMap[K, V]) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: []trace.Node{}}
	for i, p := range m.Entries() {
		s.Nodes = append(s.Nodes, trace.Node{ID: i, Label: fmt.Sprintf("%v => %v", p.Key, p.Value), Parent: -1})
		if i > 0 {
			s.Edges = append(s.Edges, trace.Edge{From: i - 1, To: i, Label: "next"})
		}
	}
	return s
}

// JSSet is a JavaScript Set: each value once, iterated in insertion order, with
// the ES2025 set methods. The zero JSSet is empty and ready.
type JSSet[T comparable] struct{ m JSMap[T, struct{}] }

// NewJSSet returns a JSSet holding values, in order.
func NewJSSet[T comparable](values ...T) *JSSet[T] {
	s := &JSSet[T]{}
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Add puts v in the set (Set.prototype.add); a repeat changes nothing.
func (s *JSSet[T]) Add(v T) *JSSet[T] {
	if !s.m.Has(v) {
		s.m.Set(v, struct{}{})
	}
	return s
}

// Has reports whether v is in the set (Set.prototype.has).
func (s *JSSet[T]) Has(v T) bool { return s.m.Has(v) }

// Delete removes v and reports whether it was there (Set.prototype.delete).
func (s *JSSet[T]) Delete(v T) bool { return s.m.Delete(v) }

// Size is the number of values (Set.prototype.size).
func (s *JSSet[T]) Size() int { return s.m.Size() }

// Values returns the values in insertion order (Set.prototype.values).
func (s *JSSet[T]) Values() []T { return s.m.Keys() }

// Union returns s's values then o's new ones (Set.prototype.union).
func (s *JSSet[T]) Union(o *JSSet[T]) *JSSet[T] {
	return NewJSSet(append(s.Values(), o.Values()...)...)
}

// Intersection returns s's values that are also in o, in s's order
// (Set.prototype.intersection).
func (s *JSSet[T]) Intersection(o *JSSet[T]) *JSSet[T] { return NewJSSet(filter(s.Values(), o.Has)...) }

// Difference returns s's values not in o (Set.prototype.difference).
func (s *JSSet[T]) Difference(o *JSSet[T]) *JSSet[T] {
	return NewJSSet(filter(s.Values(), func(v T) bool { return !o.Has(v) })...)
}

// SymmetricDifference returns the values in exactly one of s and o
// (Set.prototype.symmetricDifference).
func (s *JSSet[T]) SymmetricDifference(o *JSSet[T]) *JSSet[T] {
	return s.Difference(o).Union(o.Difference(s))
}

// IsSubsetOf reports whether every value of s is in o (Set.prototype.isSubsetOf).
func (s *JSSet[T]) IsSubsetOf(o *JSSet[T]) bool { return len(filter(s.Values(), o.Has)) == s.Size() }

// IsSupersetOf reports whether every value of o is in s.
func (s *JSSet[T]) IsSupersetOf(o *JSSet[T]) bool { return o.IsSubsetOf(s) }

// IsDisjointFrom reports whether s and o share no value.
func (s *JSSet[T]) IsDisjointFrom(o *JSSet[T]) bool { return len(filter(s.Values(), o.Has)) == 0 }

// Snapshot shows the values in insertion order.
func (s *JSSet[T]) Snapshot() trace.Snapshot {
	snap := trace.Snapshot{Layout: "set", Nodes: []trace.Node{}}
	for i, v := range s.Values() {
		snap.Nodes = append(snap.Nodes, trace.Node{ID: i, Label: fmt.Sprint(v), Parent: -1})
	}
	return snap
}
