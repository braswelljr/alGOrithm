// list - A list is a linear data structure, in which the elements are not stored at contiguous memory locations.
//
//	@function New - creates a new list
//	@function Init - initializes or clears list l.
//	@function Push - Adds a new element to the end of the list.
//	@function Pop - Removes the last element from list
//	@function Size - Returns the length of the list
//	@function Join - adds new slice with old and returns a new list
//	@function Insert - adds an element at a particular position specified if not append it add the end of the list
//	@function Remove - removes an element at a particular position specified
//	@function Get - returns an element at a particular position specified
//	@function Set - sets an element at a particular position specified
//	@function IndexOf - returns the index of an element specified
//	@function Contains - checks if an element is in the list
//	@function Clear - clears the list
//	@function Empty - checks if the list is empty
//	@function Reverse - reverses the list
//	@function Sort - sorts the list
//	@function Swap - swaps two elements in the list
//	@function String - returns a string representation of the list
//	@function Values - returns a slice of the list
//	@function Iterator - returns an iterator for the list
//	@function Equal - checks if two lists are equal
//	@function Copy - copies the list
//	@function Clone - clones the list
//	@function Merge - merges two lists
package list

// List represents a singly linked list.
type List struct {
	length int
	list   []interface{}
}

// New creates a new list.
func New(lists ...[]interface{}) *List {
	// Check for empty or null list parameter
	if lists == nil || len(lists) < 1 {
		return new(List).Init() // return a new empty list
	}

	// Create a new list
	l := &List{}

	// covert slice  to node
	// Create a new node.
	var s []interface{}

	// Iterate over the list and create a node for each value.
	for _, item := range lists {
		for _, value := range item {
			node := value
			//append slice to the end of the list
			s = append(s, node)
		}
	}
	// append slice to the end of l
	l.list = append(l.list, s...)
	// increment the length of the list
	l.length = len(l.list)
	return l
}

// Init - initializes or clears list l.
//
//	@return - *List
func (l *List) Init() *List {
	l.length = 0
	return l
}

// Push - Adds a new element to the end of the list.
//
//	@param element - element to be added
//	@return - void
func (l *List) Push(element interface{}) {
	// Create a new node.
	node := element
	// appends element to the end of the list
	l.list = append(l.list, node)
	// increment the length of the list
	l.length++
}

// Pop - Removes the last element from list
//
//	@return - void
func (l *List) Pop() {
	// slice the list from the last element
	l.list = l.list[:l.length-1]
	// decrement the length of the list
	l.length--
}

// Size - Returns the length of the list
//
//	@return - int
func (l *List) Size() int {
	// return the length of the list
	return l.length
}

// Join - adds new slice with old and returns a new list
//
//	@param slice - slice to be added
//	@return - *List
func (l *List) Join(slice []interface{}) *List {
	// covert slice  to node
	// Create a new node.
	var s []interface{}
	for _, item := range slice {
		node := item
		// append slice to the end of the list
		s = append(s, node)
	}
	// append slice to the end of l
	l.list = append(l.list, s...)
	// increment the length of the list
	l.length = len(l.list)
	return l
}

// Insert - adds an element at a particular position specified if not append it add the end of the list
//
//	@param element - element to be added
//	@param at - position of element to be added
//	@return - void
func (l *List) Insert(element interface{}, at int) {
	// set node value
	node := element
	//check for length of list
	// append item if length is zero or position is greater length
	if l.length < 1 || l.length <= at {
		// append item to the list
		l.list = append(l.list, node)
		// increase length of list
		l.length++
		return
	}

	// check for a zero or negative value(treat as zero)
	if at < 1 {
		l.list = append(append([]interface{}{}, node), l.list...)
		l.length = len(l.list)
		return
	}

	//set item position as slice index
	at--
	// make a new array and increase capacity
	// split slice at the given index
	x, y := l.list[:at], l.list[at:]

	// set list
	// append element to first slice and append second slice to the array
	l.list = append(append(x, node), y...)
	l.length = len(l.list)
}

// Remove - Remove takes out an item from list
//
//	@param element - element to be removed
//	@return - void
func (l *List) Remove(element interface{}) {
	// loop through list
	for i, item := range l.list {
		// check for an equal item
		if item == element {
			// append split slice at the given index
			l.list = append(l.list[:i], l.list[i+1:]...)
			l.length = len(l.list)
			return
		}
	}
}

// RemoveAt - takes out an item from list at a given position
//
//	@param at - position of item to be removed
//	@return - void
func (l *List) RemoveAt(at int) {
	// check for a zero or negative value(treat as zero)
	if at < 0 || at >= l.length {
		return
	}
	// loop through list
	for i := range l.list {
		// check for an equal item
		if i == at {
			// append split slice at the given index
			l.list = append(l.list[:i], l.list[i+1:]...)
			l.length = len(l.list)
			return
		}
	}
}

// Get - Returns the element at the given position
//
//	@param at - position of item to be removed
//	@return - void
func (l *List) Get(at int) interface{} {
	// check for a zero or negative value(treat as zero)
	if at < 1 {
		return nil
	}

	// set item position as slice index
	at--
	// return the element at the given position
	return l.list[at]
}

// Contains - Returns true if the list contains the element
//
//	@param element - element to be checked
//	@return - bool
func (l *List) Contains(element interface{}) bool {
	// loop through list
	for _, item := range l.list {
		// check for an equal item
		if item == element {
			return true
		}
	}
	return false
}

// Empty - empties the list
//
//	@return - void
func (l *List) Empty() {
	l.list = []interface{}{}
	l.length = len(l.list)
}

// Reverse - Reverses the list
//
//	@return - void
func (l *List) Reverse() {
	// reverse the list
	for i, j := 0, len(l.list)-1; i < j; i, j = i+1, j-1 {
		l.list[i], l.list[j] = l.list[j], l.list[i]
	}
}
