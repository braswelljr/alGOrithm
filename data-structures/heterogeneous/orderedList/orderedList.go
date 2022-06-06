package orderedList

// Node - A node in a list
// @param value - The value of the node
// @param next - The next node in the list
type Node struct {
  value interface{}
  Next  *Node
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
