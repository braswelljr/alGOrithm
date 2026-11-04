package searchtrees

import (
	"cmp"
	"fmt"
	"math/rand"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// SkipListMaxLevel is the most levels a SkipList node has.
const SkipListMaxLevel = 16

type skipNode[T cmp.Ordered] struct {
	value T
	next  []*skipNode[T]
}

// SkipList is a sorted linked list with express lanes (William Pugh, 1990):
// each node is on level 1 and, with probability 1/2, also on the next level
// up, so a search runs along the top lanes and drops down, like binary search
// on a list. Values are distinct. The zero SkipList is not ready; use
// NewSkipList.
type SkipList[T cmp.Ordered] struct {
	head  *skipNode[T]
	level int
	size  int
	rng   *rand.Rand
}

// NewSkipList returns an empty skip list whose levels come from seed, so the
// same inserts give the same shape.
func NewSkipList[T cmp.Ordered](seed int64) *SkipList[T] {
	return &SkipList[T]{head: &skipNode[T]{next: make([]*skipNode[T], SkipListMaxLevel)}, level: 1, rng: rand.New(rand.NewSource(seed))}
}

// Len is the number of values; Levels the number of levels in use. Time: O(1).
func (s *SkipList[T]) Len() int { return s.size }

// Levels is the number of levels in use.
func (s *SkipList[T]) Levels() int { return s.level }

// path returns, for each level, the last node before where v is or would be.
func (s *SkipList[T]) path(v T) []*skipNode[T] {
	update := make([]*skipNode[T], SkipListMaxLevel)
	n := s.head
	for l := s.level - 1; l >= 0; l-- {
		for n.next[l] != nil && n.next[l].value < v {
			n = n.next[l]
		}
		update[l] = n
	}
	return update
}

// Contains reports whether v is in the list. Time: O(log n) expected.
func (s *SkipList[T]) Contains(v T) bool {
	next := s.path(v)[0].next[0]
	return next != nil && next.value == v
}

// Insert adds v and reports whether it was new. Time: O(log n) expected.
func (s *SkipList[T]) Insert(v T) bool {
	update := s.path(v)
	if next := update[0].next[0]; next != nil && next.value == v {
		return false
	}
	level := 1
	for level < SkipListMaxLevel && s.rng.Intn(2) == 0 {
		level++
	}
	for l := s.level; l < level; l++ {
		update[l] = s.head
	}
	s.level = max(s.level, level)
	n := &skipNode[T]{value: v, next: make([]*skipNode[T], level)}
	for l := 0; l < level; l++ {
		n.next[l], update[l].next[l] = update[l].next[l], n
	}
	s.size++
	return true
}

// Remove deletes v and reports whether it was there. Time: O(log n) expected.
func (s *SkipList[T]) Remove(v T) bool {
	update := s.path(v)
	n := update[0].next[0]
	if n == nil || n.value != v {
		return false
	}
	for l := range n.next {
		update[l].next[l] = n.next[l]
	}
	for s.level > 1 && s.head.next[s.level-1] == nil {
		s.level--
	}
	s.size--
	return true
}

// Values returns the values in ascending order. Time: O(n).
func (s *SkipList[T]) Values() []T {
	out := make([]T, 0, s.size)
	for n := s.head.next[0]; n != nil; n = n.next[0] {
		out = append(out, n.value)
	}
	return out
}

// Snapshot describes the bottom level in order, each node labeled with its
// value and how many levels it rises to.
func (s *SkipList[T]) Snapshot() trace.Snapshot {
	var labels []string
	for n := s.head.next[0]; n != nil; n = n.next[0] {
		labels = append(labels, fmt.Sprintf("%v (%d)", n.value, len(n.next)))
	}
	return trace.Row(labels, "next", "first", "last")
}
