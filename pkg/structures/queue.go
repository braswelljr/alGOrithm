package structures

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Queue is first in, first out: Enqueue adds at the back, Dequeue takes from
// the front. It is a Deque used from both ends; the zero Queue is empty and
// ready.
type Queue[T any] struct{ d Deque[T] }

// NewQueue returns an empty queue holding values, front first.
func NewQueue[T any](values ...T) *Queue[T] {
	q := &Queue[T]{}
	for _, v := range values {
		q.Enqueue(v)
	}
	return q
}

// Enqueue adds v at the back. Time: O(1) amortized.
func (q *Queue[T]) Enqueue(v T) { q.d.PushBack(v) }

// Dequeue removes and returns the front; ok is false when empty. Time: O(1).
func (q *Queue[T]) Dequeue() (T, bool) { return q.d.PopFront() }

// Peek returns the front without removing it; ok is false when empty.
func (q *Queue[T]) Peek() (T, bool) { return q.d.Front() }

// Len is the number of values. Time: O(1).
func (q *Queue[T]) Len() int { return q.d.Len() }

// Values returns the values front to back. Time: O(n).
func (q *Queue[T]) Values() []T { return q.d.Values() }

// Snapshot describes the queue front to back.
func (q *Queue[T]) Snapshot() trace.Snapshot { return trace.Row(q.Values(), "", "front", "back") }
