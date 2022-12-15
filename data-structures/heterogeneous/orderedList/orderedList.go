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
func (list *List) Remove(element interface{}) {
	// check list and return if length is less than one
	if len(list.list) < 1 {
		return
	}

	// loop and search list for element to be removed
	for i, node := range list.list {
		if node.value == element {
			// if the element is the head
			if i == 0 {
				// set the head to the next element
				list.head, list.list[1].next = list.list[1], list.list[2]
			} else if i == len(list.list)-1 {
				// set the tail to the previous element
				list.tail, list.list[len(list.list)-2].next = list.list[len(list.list)-2], list.list[0]
			} else {
				// set the next element to the next element
				list.list[i-1].next = list.list[i+1]
			}

			// remove node from list
			list.list = append(list.list[:i], list.list[i+1:]...)
		}
	}
}

// RemoveAt deletes the node at the given index.
func (list *List) RemoveAt(at int) {
	// check for out of bounds or valid at
	if at < 0 || at >= len(list.list) {
		return
	}

	node := list.list[at]
	if node == list.head {
		// delete head
		list.head = node.next
	} else {
		// find previous node
		prev := list.list[at-1]
		prev.next = node.next
	}
	// remove node from list
	list.list = append(list.list[:at], list.list[at+1:]...)
}

// Get returns the value of the node at the given index.
func (list *List) Get(index int) interface{} {
	// check for out of bounds or valid index
	if index < 0 || index >= len(list.list) {
		return nil
	}
	// return value of node
	return list.list[index].value
}
