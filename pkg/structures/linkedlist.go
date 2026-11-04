package structures

import "github.com/braswelljr/alGOrithm/pkg/trace"

type listNode[T any] struct {
	value T
	next  *listNode[T]
}

// LinkedList is a singly linked list: each node points to the next. It keeps
// a tail pointer, so appending is O(1). The zero LinkedList is empty and
// ready.
type LinkedList[T comparable] struct {
	head, tail *listNode[T]
	size       int
}

// NewLinkedList returns a list holding values in order.
func NewLinkedList[T comparable](values ...T) *LinkedList[T] {
	l := &LinkedList[T]{}
	for _, v := range values {
		l.Append(v)
	}
	return l
}

// Len is the number of values. Time: O(1).
func (l *LinkedList[T]) Len() int { return l.size }

// Append adds v at the tail. Time: O(1).
func (l *LinkedList[T]) Append(v T) {
	n := &listNode[T]{value: v}
	if l.tail == nil {
		l.head, l.tail = n, n
	} else {
		l.tail.next, l.tail = n, n
	}
	l.size++
}

// Prepend adds v at the head. Time: O(1).
func (l *LinkedList[T]) Prepend(v T) {
	l.head = &listNode[T]{value: v, next: l.head}
	if l.tail == nil {
		l.tail = l.head
	}
	l.size++
}

// Index returns the position of the first node holding v, or -1. Time: O(n).
func (l *LinkedList[T]) Index(v T) int {
	i := 0
	for n := l.head; n != nil; n = n.next {
		if n.value == v {
			return i
		}
		i++
	}
	return -1
}

// Remove deletes the first node holding v and reports whether there was one.
// Time: O(n).
func (l *LinkedList[T]) Remove(v T) bool {
	var prev *listNode[T]
	for n := l.head; n != nil; prev, n = n, n.next {
		if n.value != v {
			continue
		}
		if prev == nil {
			l.head = n.next
		} else {
			prev.next = n.next
		}
		if n == l.tail {
			l.tail = prev
		}
		l.size--
		return true
	}
	return false
}

// Reverse reverses the list in place. Time: O(n).
func (l *LinkedList[T]) Reverse() {
	var prev *listNode[T]
	l.tail = l.head
	for n := l.head; n != nil; {
		next := n.next
		n.next = prev
		prev, n = n, next
	}
	l.head = prev
}

// Head returns the first value; ok is false when empty.
func (l *LinkedList[T]) Head() (v T, ok bool) {
	if l.head == nil {
		return v, false
	}
	return l.head.value, true
}

// Tail returns the last value; ok is false when empty.
func (l *LinkedList[T]) Tail() (v T, ok bool) {
	if l.tail == nil {
		return v, false
	}
	return l.tail.value, true
}

// Values returns the values head to tail. Time: O(n).
func (l *LinkedList[T]) Values() []T {
	out := make([]T, 0, l.size)
	for n := l.head; n != nil; n = n.next {
		out = append(out, n.value)
	}
	return out
}

// Snapshot describes the list head to tail, each node linked to the next.
func (l *LinkedList[T]) Snapshot() trace.Snapshot {
	return trace.Row(l.Values(), "next", "head", "tail")
}
