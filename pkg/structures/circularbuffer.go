package structures

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// CircularBuffer is a FIFO in a fixed array that wraps around (a ring
// buffer): no allocation after it is made, which is why audio, logging and
// network code use it. When it is full, Push either overwrites the oldest
// value or refuses, as chosen at creation. The zero CircularBuffer is not
// ready; use NewCircularBuffer.
type CircularBuffer[T any] struct {
	buf         []T
	head, size  int
	overwrite   bool
	overwritten int
}

// NewCircularBuffer returns an empty buffer of capacity slots (at least 1).
func NewCircularBuffer[T any](capacity int, overwrite bool) *CircularBuffer[T] {
	return &CircularBuffer[T]{buf: make([]T, max(1, capacity)), overwrite: overwrite}
}

// Len is the number of values; Cap the slots. Time: O(1).
func (b *CircularBuffer[T]) Len() int { return b.size }

// Cap is the number of slots.
func (b *CircularBuffer[T]) Cap() int { return len(b.buf) }

// Full reports whether every slot holds a value.
func (b *CircularBuffer[T]) Full() bool { return b.size == len(b.buf) }

// Overwritten counts the values lost to overwriting so far.
func (b *CircularBuffer[T]) Overwritten() int { return b.overwritten }

// Push adds v at the back. When the buffer is full it overwrites the oldest
// value if the buffer was made to, and otherwise refuses and returns false.
// Time: O(1).
func (b *CircularBuffer[T]) Push(v T) bool {
	if b.Full() {
		if !b.overwrite {
			return false
		}
		b.buf[b.head] = v
		b.head = (b.head + 1) % len(b.buf)
		b.overwritten++
		return true
	}
	b.buf[(b.head+b.size)%len(b.buf)] = v
	b.size++
	return true
}

// Pop removes and returns the oldest value; ok is false when empty. Time: O(1).
func (b *CircularBuffer[T]) Pop() (v T, ok bool) {
	if b.size == 0 {
		return v, false
	}
	v = b.buf[b.head]
	var zero T
	b.buf[b.head] = zero
	b.head = (b.head + 1) % len(b.buf)
	b.size--
	return v, true
}

// Peek returns the oldest value without removing it; ok is false when empty.
func (b *CircularBuffer[T]) Peek() (v T, ok bool) {
	if b.size == 0 {
		return v, false
	}
	return b.buf[b.head], true
}

// Values returns the values oldest first. Time: O(n).
func (b *CircularBuffer[T]) Values() []T {
	out := make([]T, b.size)
	for i := range out {
		out[i] = b.buf[(b.head+i)%len(b.buf)]
	}
	return out
}

// Snapshot describes the slots in memory order, so the wrap-around shows:
// the oldest value has the Role "head", the newest "tail", empty slots
// "free".
func (b *CircularBuffer[T]) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, len(b.buf))}
	for i := range b.buf {
		s.Nodes[i] = trace.Node{ID: i, Label: "", Parent: -1, Role: "free"}
	}
	for k := 0; k < b.size; k++ {
		i := (b.head + k) % len(b.buf)
		s.Nodes[i].Label, s.Nodes[i].Role = fmt.Sprint(b.buf[i]), ""
	}
	if b.size > 0 {
		s.Nodes[(b.head+b.size-1)%len(b.buf)].Role = "tail"
		s.Nodes[b.head].Role = "head"
	}
	return s
}
