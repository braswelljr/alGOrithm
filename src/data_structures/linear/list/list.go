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

// Init initializes or clears list l.
func (l *List) Init() *List {
  l.length = 0
  return l
}

// Push : Adds a new element to the end of the list.
func (l *List) Push(element interface{}) {
  // Create a new node.
  node := element
  // appends element to the end of the list
  l.list = append(l.list, node)
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

// Join : Joins new slice with old and returns a new list
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

//Insert : Inserts an element at a particular position specified if not append it add the end of the list
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

// Remove : Remove takes out an item from list
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

// RemoveAt : RemoveAt takes out an item from list at a given position
func (l *List) RemoveAt(at int) {
  // check for a zero or negative value(treat as zero)
  if at < 1 || at > l.length {
    return
  }
  // loop through list
  for i, _ := range l.list {
    // check for an equal item
    if i == at {
      // append split slice at the given index
      l.list = append(l.list[:i], l.list[i+1:]...)
      l.length = len(l.list)
      return
    }
  }
}

// Get : Returns the element at the given position
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

// Contains : Returns true if the list contains the element
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

// Clear : Clears the list
func (l *List) Clear() {
  l.list = []interface{}{}
  l.length = len(l.list)
}
