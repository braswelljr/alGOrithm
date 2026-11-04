package structures

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MonotonicQueue is a FIFO that also answers "what is the largest value in the
// queue?" in O(1) amortized, the trick behind the sliding-window maximum:
// beside the values it keeps a deque of the candidates for the maximum, in
// decreasing order; a new value removes every smaller candidate, since none
// of them can be the maximum while it is in the queue. The zero
// MonotonicQueue is empty and ready.
type MonotonicQueue[T cmp.Ordered] struct {
	values []T
	front  int   // values[front:] is the queue
	maxes  []int // absolute indexes of the candidates, values decreasing
}

// NewMonotonicQueue returns a queue holding values, front first.
func NewMonotonicQueue[T cmp.Ordered](values ...T) *MonotonicQueue[T] {
	q := &MonotonicQueue[T]{}
	for _, v := range values {
		q.Enqueue(v)
	}
	return q
}

// Enqueue adds v at the back. Time: O(1) amortized.
func (q *MonotonicQueue[T]) Enqueue(v T) {
	for len(q.maxes) > 0 && q.values[q.maxes[len(q.maxes)-1]] < v {
		q.maxes = q.maxes[:len(q.maxes)-1]
	}
	q.values = append(q.values, v)
	q.maxes = append(q.maxes, len(q.values)-1)
}

// Dequeue removes and returns the front; ok is false when empty.
// Time: O(1) amortized.
func (q *MonotonicQueue[T]) Dequeue() (v T, ok bool) {
	if q.Len() == 0 {
		return v, false
	}
	v = q.values[q.front]
	if q.maxes[0] == q.front {
		q.maxes = q.maxes[1:]
	}
	q.front++
	if q.front > 32 && q.front > len(q.values)/2 { // reclaim the space in front
		q.values = append([]T(nil), q.values[q.front:]...)
		for i := range q.maxes {
			q.maxes[i] -= q.front
		}
		q.front = 0
	}
	return v, true
}

// Peek returns the front without removing it; ok is false when empty.
func (q *MonotonicQueue[T]) Peek() (v T, ok bool) {
	if q.Len() == 0 {
		return v, false
	}
	return q.values[q.front], true
}

// Max returns the largest value in the queue; ok is false when empty.
// Time: O(1).
func (q *MonotonicQueue[T]) Max() (v T, ok bool) {
	if q.Len() == 0 {
		return v, false
	}
	return q.values[q.maxes[0]], true
}

// Len is the number of values. Time: O(1).
func (q *MonotonicQueue[T]) Len() int { return len(q.values) - q.front }

// Snapshot describes the queue front to back; the candidates for the maximum
// have the Role "candidate", and the maximum "max".
func (q *MonotonicQueue[T]) Snapshot() trace.Snapshot {
	s := trace.Row(q.values[q.front:], "", "", "")
	for i, idx := range q.maxes {
		role := "candidate"
		if i == 0 {
			role = "max"
		}
		s.Nodes[idx-q.front].Role = role
	}
	return s
}

var _ FIFO[int] = (*MonotonicQueue[int])(nil)
