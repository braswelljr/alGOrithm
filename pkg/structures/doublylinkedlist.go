package structures

import "github.com/braswelljr/alGOrithm/pkg/trace"

type dNode[T any] struct {
	value      T
	prev, next *dNode[T]
}

// DoublyLinkedList is a list whose nodes point both ways, so it can be
// walked, and changed at either end, in O(1). The zero list is empty and
// ready.
type DoublyLinkedList[T comparable] struct {
	head, tail *dNode[T]
	size       int
}

// NewDoublyLinkedList returns a list holding values in order.
func NewDoublyLinkedList[T comparable](values ...T) *DoublyLinkedList[T] {
	l := &DoublyLinkedList[T]{}
	for _, v := range values {
		l.PushBack(v)
	}
	return l
}

// Len is the number of values. Time: O(1).
func (l *DoublyLinkedList[T]) Len() int { return l.size }

// PushBack adds v at the tail. Time: O(1).
func (l *DoublyLinkedList[T]) PushBack(v T) {
	n := &dNode[T]{value: v, prev: l.tail}
	if l.tail == nil {
		l.head = n
	} else {
		l.tail.next = n
	}
	l.tail = n
	l.size++
}

// PushFront adds v at the head. Time: O(1).
func (l *DoublyLinkedList[T]) PushFront(v T) {
	n := &dNode[T]{value: v, next: l.head}
	if l.head == nil {
		l.tail = n
	} else {
		l.head.prev = n
	}
	l.head = n
	l.size++
}

func (l *DoublyLinkedList[T]) unlink(n *dNode[T]) {
	if n.prev == nil {
		l.head = n.next
	} else {
		n.prev.next = n.next
	}
	if n.next == nil {
		l.tail = n.prev
	} else {
		n.next.prev = n.prev
	}
	l.size--
}

// PopFront removes and returns the head; ok is false when empty. Time: O(1).
func (l *DoublyLinkedList[T]) PopFront() (v T, ok bool) {
	if l.head == nil {
		return v, false
	}
	n := l.head
	l.unlink(n)
	return n.value, true
}

// PopBack removes and returns the tail; ok is false when empty. Time: O(1).
func (l *DoublyLinkedList[T]) PopBack() (v T, ok bool) {
	if l.tail == nil {
		return v, false
	}
	n := l.tail
	l.unlink(n)
	return n.value, true
}

// Remove deletes the first node holding v and reports whether there was one.
// Time: O(n) to find it, O(1) to unlink it.
func (l *DoublyLinkedList[T]) Remove(v T) bool {
	for n := l.head; n != nil; n = n.next {
		if n.value == v {
			l.unlink(n)
			return true
		}
	}
	return false
}

// Values returns the values head to tail. Time: O(n).
func (l *DoublyLinkedList[T]) Values() []T {
	out := make([]T, 0, l.size)
	for n := l.head; n != nil; n = n.next {
		out = append(out, n.value)
	}
	return out
}

// Backward returns the values tail to head, walking the prev links. Time: O(n).
func (l *DoublyLinkedList[T]) Backward() []T {
	out := make([]T, 0, l.size)
	for n := l.tail; n != nil; n = n.prev {
		out = append(out, n.value)
	}
	return out
}

// Snapshot describes the list head to tail with "next" and "prev" edges.
func (l *DoublyLinkedList[T]) Snapshot() trace.Snapshot {
	s := trace.Row(l.Values(), "next", "head", "tail")
	for i := 1; i < len(s.Nodes); i++ {
		s.Edges = append(s.Edges, trace.Edge{From: i, To: i - 1, Label: "prev"})
	}
	return s
}
