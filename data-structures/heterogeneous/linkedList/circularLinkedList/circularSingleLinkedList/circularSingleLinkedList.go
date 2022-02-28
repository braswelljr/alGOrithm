package circularSingleLinkedList

// Node is a node of a single linked list.
type Node struct {
  value interface{}
  next  *Node
}

type List struct {
  head     *Node
  tail     *Node
  elements []*Node
}

// New returns a new single linked list.
func New() *List {
  return &List{}
}

// Init initializes a single linked list.
func (list *List) Init() {
  list.head = nil
  list.tail = nil
  list.elements = []*Node{}
}

// Insert inserts a new node at the end of the list.
func (list *List) Insert(element interface{}) {
  // create a new node
  node := &Node{value: element}
  if len(list.elements) < 1 || list.head == nil {
    // If the list is empty, set the head and tail to the new node.
    list.head, list.tail = node, node
  } else {
    // set next pointer of node to head
    node.next = list.head
    // If the list is not empty, set the next pointer of the tail to the new node.
    list.tail, list.tail.next, list.elements[len(list.elements)-1].next = node, list.head, node
  }

  // append node to elements
  list.elements = append(list.elements, node)
}

// Head returns the head of the list.
func (list *List) Head() interface{} {
  // return nil if list is empty
  if len(list.elements) < 1 || list.head == nil {
    return nil
  }
  // return head value
  return list.head.value
}

// Tail returns the tail of the list.
func (list *List) Tail() interface{} {
  // return nil if list is empty
  if len(list.elements) < 1 || list.tail == nil {
    return nil
  }
  // return tail value
  return list.tail.value
}

// Empty sets the list to empty
func (list *List) Empty() {
  // set head and tail to nil
  list.head = nil
  list.tail = nil
  // set elements to empty slice
  list.elements = []*Node{}
}

// Values returns a slice of the values in the list.
func (list *List) Values() []interface{} {
  // create slice to hold values
  values := make([]interface{}, len(list.elements))
  // iterate over list
  for i, node := range list.elements {
    // add value to slice
    values[i] = node.value
  }
  // return slice
  return values
}
