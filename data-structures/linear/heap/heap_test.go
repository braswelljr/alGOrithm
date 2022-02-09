package heap

import "testing"

// TestHeap tests the heap implementation.
func TestHeap(t *testing.T) {
	h := New([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	h.Push(1)
	h.Push(2)
	h.Push(16)
	h.Push(17)
	t.Log("heap:", h)
}
