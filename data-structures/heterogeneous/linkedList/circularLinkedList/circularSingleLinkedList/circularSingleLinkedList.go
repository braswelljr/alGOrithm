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
		// If the list is not empty, set next pointer of node to head and set the next pointer of the tail to the new node.
		node.next, list.tail, list.tail.next, list.elements[len(list.elements)-1].next = list.head, node, list.head, node
	}

	// append node to elements
	list.elements = append(list.elements, node)
}

// Remove deletes the given node from the list.
func (list *List) Remove(element interface{}) {
	// If the list is empty, return.
	if len(list.elements) < 1 {
		return
	}

	// loop through elements and find the element to remove
	for i, node := range list.elements {
		// check if the element is the one to remove
		if node.value == element {
			// if the element is the head
			if i == 0 {
				// set the head to the next element
				list.head = list.elements[1]
			} else if i == len(list.elements)-1 {
				// set the tail to the previous element
				list.tail = list.elements[len(list.elements)-2]
			} else {
				// set the next element of the previous element to the next element
				list.elements[i-1].next = list.elements[i+1]
			}

			// remove node from list
			list.elements = append(list.elements[:i], list.elements[i+1:]...)
		}
	}
}

// RemoveAt deletes the node at the given index.
func (list *List) RemoveAt(at int) {
	// check for out of bounds or valid at
	if at < 0 || at >= len(list.elements) {
		return
	}

	node := list.elements[at]
	// if the element is the head
	if node == list.elements[0] {
		// set the head to the next element
		list.head = list.elements[1]
	} else if node == list.elements[len(list.elements)-1] {
		// set the tail to the previous element
		list.tail = list.elements[len(list.elements)-2]
	} else {
		// set the next element of the previous element to the next element
		list.elements[at-1].next = list.elements[at+1]
	}
	// remove node from list
	list.elements = append(list.elements[:at], list.elements[at+1:]...)
}

// Get returns the value of the node at the given index.
func (list *List) Get(index int) interface{} {
	// check for out of bounds or valid index
	if index < 0 || index >= len(list.elements) {
		return nil
	}
	// return value of node
	return list.elements[index].value
}

// Set sets the value of the node at the given index.
func (list *List) Set(index int, value interface{}) {
	// check for out of bounds or valid index
	if index < 0 || index >= len(list.elements) {
		return
	}
	// set value of node
	list.elements[index].value = value
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
		if node.value == value {
			return i
		}
	}
	// return -1 if not found
	return -1
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
