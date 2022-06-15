package orderedList

// Node - A node in a list
// @param value - The value of the node
// @param next - The next node in the list
type Node struct {
  value interface{}
  next  *Node
}

// List - A list of nodes
// @param head - The first node in the list
// @param tail - The last node in the list
// @param list - The list of nodes
type List struct {
  head *Node
  tail *Node
  list []*Node
}

// New returns a new double linked list.
func New() *List {
  return &List{}
}

// Init initializes a double linked list with the values in the slice.
func (list *List) Init() {
  list.head = nil
  list.tail = nil
  list.list = []*Node{}
}

// Insert - adds a new item to the list
func (list *List) Insert(value interface{}) {
  // create a new node
  node := &Node{value: value}

  // set head and tail to new node if list was empty
  if len(list.list) < 1 {
    list.head, list.tail = node, node
  } else {
    // set the next of the last node of list and tail to new node
    list.list[len(list.list)-1].next, list.tail = node, node
  }

  // append node to list
  list.list = append(list.list, node)

  // set the next value of the tail to the head
  list.tail.next = list.head
}

// Remove removes the node from the list.
func (list *List) Remove(element interface{}) {}
