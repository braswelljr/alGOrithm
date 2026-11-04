package structures

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Deque is a double-ended queue on a ring buffer: push and pop at either end
// in O(1), doubling when full. The zero Deque is empty and ready.
type Deque[T any] struct {
	buf        []T
	head, size int
}

// NewDeque returns an empty deque.
func NewDeque[T any]() *Deque[T] { return &Deque[T]{} }

func (d *Deque[T]) grow() {
	if d.size < len(d.buf) {
		return
	}
	buf := make([]T, max(4, 2*len(d.buf)))
	for i := 0; i < d.size; i++ {
		buf[i] = d.buf[(d.head+i)%len(d.buf)]
	}
	d.buf, d.head = buf, 0
}

// Len is the number of values. Time: O(1).
func (d *Deque[T]) Len() int { return d.size }

// PushFront adds v at the front. Time: O(1) amortized.
func (d *Deque[T]) PushFront(v T) {
	d.grow()
	d.head = (d.head - 1 + len(d.buf)) % len(d.buf)
	d.buf[d.head] = v
	d.size++
}

// PushBack adds v at the back. Time: O(1) amortized.
func (d *Deque[T]) PushBack(v T) {
	d.grow()
	d.buf[(d.head+d.size)%len(d.buf)] = v
	d.size++
}

// PopFront removes and returns the front; ok is false when empty. Time: O(1).
func (d *Deque[T]) PopFront() (v T, ok bool) {
	if d.size == 0 {
		return v, false
	}
	v = d.buf[d.head]
	var zero T
	d.buf[d.head] = zero
	d.head = (d.head + 1) % len(d.buf)
	d.size--
	return v, true
}

// PopBack removes and returns the back; ok is false when empty. Time: O(1).
func (d *Deque[T]) PopBack() (v T, ok bool) {
	if d.size == 0 {
		return v, false
	}
	i := (d.head + d.size - 1) % len(d.buf)
	v = d.buf[i]
	var zero T
	d.buf[i] = zero
	d.size--
	return v, true
}

// Front returns the front without removing it; ok is false when empty.
func (d *Deque[T]) Front() (v T, ok bool) {
	if d.size == 0 {
		return v, false
	}
	return d.buf[d.head], true
}

// Back returns the back without removing it; ok is false when empty.
func (d *Deque[T]) Back() (v T, ok bool) {
	if d.size == 0 {
		return v, false
	}
	return d.buf[(d.head+d.size-1)%len(d.buf)], true
}

// Values returns the values front to back. Time: O(n).
func (d *Deque[T]) Values() []T {
	out := make([]T, d.size)
	for i := range out {
		out[i] = d.buf[(d.head+i)%len(d.buf)]
	}
	return out
}

// Snapshot describes the deque front to back.
func (d *Deque[T]) Snapshot() trace.Snapshot { return trace.Row(d.Values(), "", "front", "back") }
