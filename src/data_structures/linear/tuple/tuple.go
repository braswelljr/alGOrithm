package tuple

import "strings"

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

// ToString returns a string representation of the tuple
func (tuple *Tuple) ToString() string {
  // create a string builder
  var builder strings.Builder
  // append the opening bracket
  builder.WriteString("[")
  // iterate over the list and append the elements
  for index, item := range tuple.elements {
    // append the element
    builder.WriteString(item.(string))
    // append the comma if the element is not the last element
    if index < len(tuple.elements)-1 {
      builder.WriteString(", ")
    }
  }
  // append the closing bracket
  builder.WriteString("]")
  // return the string representation of the list
  return builder.String()
}
