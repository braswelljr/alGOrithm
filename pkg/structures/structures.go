// Package structures holds the linear data structures and the access
// principles, one file each: Stack (LIFO), Queue (FIFO), Deque, LinkedList,
// DoublyLinkedList, CircularBuffer, MinStack, MonotonicQueue and SortedList,
// with the LIFO and FIFO interfaces in principles.go. All are generic.
//
// The other structures are in subpackages, one per family: heaps,
// searchtrees, ranges, lookup, probabilistic, spatial, persistent and
// buffers, beside arrays, sets, trees and collections.
//
// Anything that can find nothing returns (value, ok) instead of a zero value
// or a panic. Every structure the visualizer draws has a Snapshot method that
// describes its shape as nodes and edges (a trace.Snapshot):
//
//	s := structures.NewStack[int]()
//	s.Push(1)
//	if v, ok := s.Pop(); ok { ... }
package structures
