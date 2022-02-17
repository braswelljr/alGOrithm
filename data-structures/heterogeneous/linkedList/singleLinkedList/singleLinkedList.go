package singleLinkedList

// Node is a node of a single linked list.
type Node struct {
  Value interface{}
  Next  *Node
}

type List struct {
  Head     *Node
  Tail     *Node
  elements []*Node
}

// New returns a new single linked list.
func New() *List {
  return &List{}
}

// Init initializes a single linked list.
func (list *List) Init() {
  list.Head = nil
  list.Tail = nil
  list.elements = []*Node{}
}

// Insert inserts a new node at the end of the list.
func (list *List) Insert(value interface{}) {
  // Create a new node.
  node := &Node{Value: value}
  if list.Head == nil {
    // If the list is empty, set the head and tail to the new node.
    list.Head = node
    list.Tail = node
  } else {
    // If the list is not empty, set the next pointer of the tail to the new node.
    list.Tail.Next = node
    list.Tail = node
  }
  // append node to elements
  list.elements = append(list.elements, node)
}

// Remove deletes the node at the given index.
func (list *List) Remove(index int) {
  // check for out of bounds or valid index
  if index < 0 || index >= len(list.elements) {
    return
  }

  node := list.elements[index]
  if node == list.Head {
    // delete head
    list.Head = node.Next
  } else {
    // find previous node
    prev := list.elements[index-1]
    prev.Next = node.Next
  }
  // remove node from list
  list.elements = append(list.elements[:index], list.elements[index+1:]...)
}

// Get returns the value of the node at the given index.
func (list *List) Get(index int) interface{} {
  // check for out of bounds or valid index
  if index < 0 || index >= len(list.elements) {
    return nil
  }
  // return value of node
  return list.elements[index].Value
}

// Set sets the value of the node at the given index.
func (list *List) Set(index int, value interface{}) {
  // check for out of bounds or valid index
  if index < 0 || index >= len(list.elements) {
    return
  }
  // set value of node
  list.elements[index].Value = value
}

// Length returns the length of the list.
func (list *List) Length() int {
  return len(list.elements)
}

// Search searches the list for the given value.
func (list *List) Search(value interface{}) int {
  // iterate over list
  for i, node := range list.elements {
    // check if value matches
    if node.Value == value {
      return i
    }
  }
  // return -1 if not found
  return -1
}

// Reverse reverses the list.
func (list *List) Reverse() {
  // iterate over list
  for i := 0; i < len(list.elements)/2; i++ {
    // swap nodes
    list.elements[i], list.elements[len(list.elements)-i-1] = list.elements[len(list.elements)-i-1], list.elements[i]
  }
}
