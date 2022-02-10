package tuple

import (
	"strconv"
)

// Tuple is a finite ordered list (sequence) of elements.
// Tuples are immutable forms of Lists.
// Tuple is a reference type.

type Tuple struct {
	elements []interface{}
}

// Init initializes a new tuple
func (tuple *Tuple) Init() *Tuple {
	return &Tuple{}
}

// New calls the Init function to initialize a new Tuple
func New(elements ...[]interface{}) *Tuple {
	if elements == nil || len(elements) < 1 {
		return new(Tuple).Init()
	}

	// covert slice  to node
	// Create a new node.
	var element []interface{}

	// Iterate over the list and create a node for each value.
	for _, item := range elements {
		for _, value := range item {
			node := value
			//append slice to the end of the list
			element = append(element, node)
		}
	}

	// Create a new node.
	return &Tuple{element}
}

// Size returns the number of elements in the tuple
func (tuple *Tuple) Size() int {
	return len(tuple.elements)
}

// Get returns the element at the given index
func (tuple *Tuple) Get(index int) interface{} {
	// check if index is within bounds
	if index < 0 || index >= len(tuple.elements) {
		return nil
	}
	// return element
	return tuple.elements[index]
}

// Set sets the element at the given index
func (tuple *Tuple) Set(index int, element interface{}) *Tuple {
	// check if index is within bounds
	if index < 0 || index >= len(tuple.elements) {
		return tuple
	}
	// set element
	tuple.elements[index] = element
	// return tuple
	return tuple
}

// Contains returns true if the tuple contains the given element
func (tuple *Tuple) Contains(element interface{}) bool {
	// iterate over the list and return true if the element is found
	for _, item := range tuple.elements {
		if item == element {
			return true
		}
	}
	// return false if the element is not found
	return false
}

// IndexOf returns the index of the given element
func (tuple *Tuple) IndexOf(element interface{}) int {
	// iterate over the list and return the index if the element is found
	for index, item := range tuple.elements {
		if item == element {
			return index
		}
	}
	// return -1 if the element is not found
	return -1
}

// ToSlice returns the elements of the tuple as a slice
func (tuple *Tuple) ToSlice() []interface{} {
	return tuple.elements
}

// Equals returns true if the tuple is equal to the given tuple
func (tuple *Tuple) Equals(other *Tuple) bool {
	// check if the other tuple is nil
	if other == nil {
		return false
	}
	// check if the other tuple is equal to the current tuple
	if tuple.Size() != other.Size() {
		return false
	}
	// iterate over the list and return false if the element is not equal
	for index, item := range tuple.elements {
		if item != other.Get(index) {
			return false
		}
	}
	// return true if all elements are equal
	return true
}

// Clone returns a clone of the tuple
func (tuple *Tuple) Clone() *Tuple {
	// create a new tuple
	clone := new(Tuple).Init()
	// iterate over the list and append the elements
	for _, item := range tuple.elements {
		clone.elements = append(clone.elements, item)
	}
	// return the clone
	return clone
}

// ToMap returns the elements of the tuple as a map
func (tuple *Tuple) ToMap() map[string]interface{} {
	// create a new map
	mapTuple := make(map[string]interface{})
	// iterate over the list and append the elements
	for index, item := range tuple.elements {
		mapTuple[strconv.Itoa(index)] = item
	}
	// return the map
	return mapTuple
}

// Reverse returns a new tuple with the elements of the tuple in reverse order
func (tuple *Tuple) Reverse() *Tuple {
	// create a new tuple
	reverse := new(Tuple).Init()
	// iterate over the list and append the elements
	for index := len(tuple.elements) - 1; index >= 0; index-- {
		reverse.elements = append(reverse.elements, tuple.elements[index])
	}
	// return the reversed tuple
	return reverse
}

// ToArray returns the elements of the tuple as an array
func (tuple *Tuple) ToArray() []interface{} {
	// create a new array
	arrayTuple := make([]interface{}, tuple.Size())
	// iterate over the list and append the elements
	for index, item := range tuple.elements {
		arrayTuple[index] = item
	}
	// return the array
	return arrayTuple
}
