package list

// List represents a singly linked list.
type List struct {
	length int
	list   []interface{}
}

// New creates a new list.
// Convert params(i.e. lists) to []interface{} with @func TypeToInterface.
func New(lists ...[]interface{}) *List {
	// Create a new list.
	// Check for empty or null list parameter
	if lists == nil || len(lists) == 0 {
		return new(List).Init() // return a new empty list
	}

	// Create a new list.
	var l []interface{}
	// Check for one or more list parameters and covert to one list
	if len(lists) == 1 {
		l = lists[0]
	} else {
		for _, list := range lists {
			l = append(list, list...)
		}
	}

	return &List{
		length: len(l),
		list:   l,
	}
}

// Init initializes or clears list l.
func (l *List) Init() *List {
	l.length = 0
	return l
}

// Push : Adds a new element to the end of the list.
func (l *List) Push(element interface{}) {
	// appends element to the end of the list
	l.list = append(l.list, element)
	// increment the length of the list
	l.length++
}

// Pop : Removes the last element from list
func (l *List) Pop() {
	// slice the list from the last element
	l.list = l.list[:l.length-1]
	// decrement the length of the list
	l.length--
}

// Size : Returns the length of the list
func (l *List) Size() int {
	// return the length of the list
	return l.length
}

