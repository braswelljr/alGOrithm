package doubleLinkedList

// Node represents a node in a double linked list.
type Node struct {
	value interface{}
	prev  *Node
	next  *Node
}

// List represents a double linked list.
type List struct {
	head     *Node
	tail     *Node
	elements []*Node
}

// New returns a new double linked list.
func New() *List {
	return &List{}
}

// Init initializes a double linked list with the values in the slice.
func (list *List) Init(values []interface{}) {
	list.head = nil
	list.tail = nil
	list.elements = []*Node{}
}

// Insert inserts a new node with the given value at the end of the list.
func (list *List) Insert(value interface{}) {
	// create a new node
	node := &Node{value: value}
	if list.head == nil {
		// if the list is empty, set the head and tail to the new node
		list.head = node
		list.tail = node
	} else {
		// otherwise, set the new node's next to the current tail and the current tail's prev to the new node
		list.tail.next = node
		node.prev = list.tail
		list.tail = node
	}
	// add the new node to the list
	list.elements = append(list.elements, node)
}

// Remove removes the node from the list.
func (list *List) Remove(element interface{}) {
	// check length of list
	if len(list.elements) < 1 {
		return
	}
	// find the node
	for i, node := range list.elements {
		if node.value == element {
			// if the node is the head, set the head to the next node
			if node == list.head {
				list.head = node.next
			}
			// if the node is the tail, set the tail to the prev node
			if node == list.tail {
				list.tail = node.prev
			}
			// if the node has a prev, set the prev's next to the node's next
			if node.prev != nil {
				node.prev.next = node.next
			}
			// if the node has a next, set the next's prev to the node's prev
			if node.next != nil {
				node.next.prev = node.prev
			}
			// remove the node from the list
			list.elements = append(list.elements[:i], list.elements[i+1:]...)
			break
		}
	}
}

// RemoveAt removes the node at the given index from the list.
func (list *List) RemoveAt(at int) {
	// if the at is out of bounds, return
	if at < 0 || at >= len(list.elements) {
		return
	}
	// get the node at the given at
	node := list.elements[at]
	// if the node is the head, set the head to the next node
	if node == list.head {
		list.head = node.next
	}
	// if the node is the tail, set the tail to the prev node
	if node == list.tail {
		list.tail = node.prev
	}
	// if the node has a prev node, set the prev node's next to the node's next
	if node.prev != nil {
		node.prev.next = node.next
	}
	// if the node has a next node, set the next node's prev to the node's prev
	if node.next != nil {
		node.next.prev = node.prev
	}
	// remove the node from the list
	list.elements = append(list.elements[:at], list.elements[at+1:]...)
}

// Search searches the list for the given value and returns the at of the first occurrence.
func (list *List) Search(value interface{}) int {
	// iterate over the list
	for i, node := range list.elements {
		// if the node's value is the given value, return the at
		if node.value == value {
			return i
		}
	}
	// if the value is not found, return -1
	return -1
}

// Reverse reverses the list.
func (list *List) Reverse() {
	// iterate over the list
	for i := 0; i < len(list.elements)/2; i++ {
		// get the node at the current index
		node, rev := list.elements[i], list.elements[len(list.elements)-1-i]
		// set the node at the current index's next to the node at the length - 1 - index's next
		node.next = rev.next
		// set the node at the length - 1 - index's next to the node at the current index
		rev.next = node
		// set the node at the current index's prev to the node at the length - 1 - index's prev
		node.prev = rev.prev
		// set the node at the length - 1 - index's prev to the node at the current index
		rev.prev = node
	}
	// set the head and tail to the new head and tail
	list.head, list.tail = list.tail, list.head
}

// Get returns the value at the given index.
func (list *List) Get(index int) interface{} {
	// if the index is out of bounds, return nil
	if index < 0 || index >= len(list.elements) {
		return nil
	}
	// return the value at the given index
	return list.elements[index].value
}

// Set sets the value at the given index.
func (list *List) Set(index int, value interface{}) {
	// if the index is out of bounds, return
	if index < 0 || index >= len(list.elements) {
		return
	}
	// set the value at the given index
	list.elements[index].value = value
}

// Length returns the length of the list.
func (list *List) Length() int {
	// return the length of the list
	return len(list.elements)
}
