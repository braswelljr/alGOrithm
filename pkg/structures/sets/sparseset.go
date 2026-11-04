package sets

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// SparseSet holds integers 0..universe-1 in two arrays (Preston Briggs and
// Linda Torczon, 1993): dense lists the members in insertion order and
// sparse[v] says where v is in dense. v is a member when sparse[v] points
// inside dense and dense points back at v, so neither array ever needs to be
// cleared: Clear is O(1) and iterating visits only the members, where a
// Bitset would scan the whole universe. Compilers and game engines use it for
// sets that are emptied and refilled constantly.
//
// Time: O(1) Add, Remove, Contains and Clear; O(n) to list n members.
// Space: O(universe).
type SparseSet struct {
	dense  []int
	sparse []int
}

// NewSparseSet returns an empty set over 0..universe-1 (universe at least 1).
func NewSparseSet(universe int) *SparseSet {
	return &SparseSet{sparse: make([]int, max(universe, 1))}
}

// Universe returns how many integers the set can hold.
func (s *SparseSet) Universe() int { return len(s.sparse) }

// Contains reports whether v is a member.
func (s *SparseSet) Contains(v int) bool {
	if v < 0 || v >= len(s.sparse) {
		return false
	}
	i := s.sparse[v]
	return i < len(s.dense) && s.dense[i] == v
}

// Add adds v and reports whether it was not a member; v outside the universe
// is refused.
func (s *SparseSet) Add(v int) bool {
	if v < 0 || v >= len(s.sparse) || s.Contains(v) {
		return false
	}
	s.sparse[v] = len(s.dense)
	s.dense = append(s.dense, v)
	return true
}

// Remove deletes v and reports whether it was a member; the last member
// moves into its slot, so the order of Values changes.
func (s *SparseSet) Remove(v int) bool {
	if !s.Contains(v) {
		return false
	}
	i, last := s.sparse[v], s.dense[len(s.dense)-1]
	s.dense[i], s.sparse[last] = last, i
	s.dense = s.dense[:len(s.dense)-1]
	return true
}

// Clear empties the set in O(1), leaving the sparse array as it is.
func (s *SparseSet) Clear() { s.dense = s.dense[:0] }

// Len returns the number of members.
func (s *SparseSet) Len() int { return len(s.dense) }

// Values returns the members in dense order.
func (s *SparseSet) Values() []int { return append([]int(nil), s.dense...) }

// Snapshot describes the dense array, each member with its index into it.
func (s *SparseSet) Snapshot() trace.Snapshot {
	labels := make([]string, len(s.dense))
	for i, v := range s.dense {
		labels[i] = fmt.Sprintf("dense[%d]=%d", i, v)
	}
	return trace.Row(labels, "", "", "")
}
