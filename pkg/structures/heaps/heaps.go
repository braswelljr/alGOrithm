// Package heaps holds heaps, the structures behind priority queues: MinHeap, a
// binary heap over a slice, and PriorityHeap, which orders items by a
// separate priority and can change it in place. Both give back the most urgent
// value first, the third access principle beside LIFO and FIFO (see
// structures.LIFO and structures.FIFO).
package heaps

// PriorityQueue gives back the most urgent value first, whatever the order it
// went in; for a MinHeap the most urgent is the smallest.
type PriorityQueue[T any] interface {
	Push(v T)
	Pop() (T, bool)
	Top() (T, bool)
	Len() int
}

var _ PriorityQueue[int] = (*MinHeap[int])(nil)
