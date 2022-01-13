package heap

import (
	"fmt"
	"strings"
)

type Heap struct {
	data []int
}

// Init initializes the heap with the given data.
func (h *Heap) Init(data []int) {
	h.data = data
	for i := len(data) / 2; i >= 0; i-- {
		h.down(i)
	}
}

// New returns a new heap with the given data.
func New(data []int) *Heap {
	h := new(Heap)
	h.Init(data)
	return h
}

// Push pushes the given value into the heap.
func (h *Heap) Push(value int) {
	h.data = append(h.data, value)
	h.up(len(h.data) - 1)
}

// Pop pops the top value from the heap.
func (h *Heap) Pop() int {
	if len(h.data) == 0 {
		return 0
	}
	value := h.data[0]
	h.data[0] = h.data[len(h.data)-1]
	h.data = h.data[:len(h.data)-1]
	h.down(0)
	return value
}

// down moves the value at the given index down the heap.
func (h *Heap) down(index int) {
	for {
		left := 2*index + 1
		right := 2*index + 2
		if left >= len(h.data) {
			break
		}
		if right >= len(h.data) {
			if h.data[left] < h.data[index] {
				h.swap(index, left)
				index = left
			} else {
				break
			}
		} else {
			if h.data[left] < h.data[index] && h.data[left] < h.data[right] {
				h.swap(index, left)
				index = left
			} else if h.data[right] < h.data[index] && h.data[right] < h.data[left] {
				h.swap(index, right)
				index = right
			} else {
				break
			}
		}
	}
}

// up moves the value at the given index up the heap.
func (h *Heap) up(index int) {
	for {
		parent := (index - 1) / 2
		if parent < 0 {
			break
		}
		if h.data[parent] < h.data[index] {
			h.swap(parent, index)
			index = parent
		} else {
			break
		}
	}
}

// swap the values at the given indices.
func (h *Heap) swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

// String returns a string representation of the heap.
func (h *Heap) String() string {
	return "[" + strings.Join(strings.Fields(fmt.Sprint(h.data)), ", ") + "]"
}

// Valid Test if the heap is valid.
func (h *Heap) Valid() bool {
	for i := 0; i < len(h.data); i++ {
		left := 2*i + 1
		right := 2*i + 2
		if left < len(h.data) && h.data[left] < h.data[i] {
			return false
		}
		if right < len(h.data) && h.data[right] < h.data[i] {
			return false
		}
	}
	return true
}

// Empty returns true if the heap is empty.
func (h *Heap) Empty() bool {
	return len(h.data) == 0
}

// Size returns the number of elements in the heap.
func (h *Heap) Size() int {
	return len(h.data)
}

// Top returns the top value of the heap.
func (h *Heap) Top() int {
	if len(h.data) == 0 {
		return 0
	}
	return h.data[0]
}

// Copy returns a copy of the heap.
func (h *Heap) Copy() *Heap {
	return New(h.data)
}
