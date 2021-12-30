package list

// List represents a singly linked list.
type List struct {
  length int
  list   []Node
}

type Node struct {
  value    int
  previous *Node
  next     *Node
}

// New creates a new list.
func New(lists ...[]int) *List {
  // Check for empty or null list parameter
  if lists == nil || len(lists) < 1 {
    return new(List).Init() // return a new empty list
  }

  // node
  var node *Node
  var l []Node

  // not an empty list
  for i, list := range lists {
    node = &Node{
      value:    list[i],
      previous: nil,
      next:     nil,
    }

    //add node to list
    l = append(l, *node)
  }

  // Create a new list.
  list := append([]Node{}, l...)

  return &List{
    length: len(list),
    list:   list,
  }
}

// Init initializes or clears list l.
func (l *List) Init() *List {
  l.length = 0
  return l
}

// Push : Adds a new element to the end of the list.
func (l *List) Push(element int) {
  // Create a new node.
  node := &Node{
    value: element,
  }
  // appends element to the end of the list
  l.list = append(l.list, *node)
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
func (l *List) Join(slice []int) *List {
  // covert slice  to node
  // Create a new node.
  var s []Node
  for _, item := range slice {
    node := &Node{
      value: item,
    }
    // append slice to the end of the list
    s = append(l.list, *node)
  }
  // append slice to the end of l
  l.list = append(l.list, s...)
  // increment the length of the list
  l.length = len(l.list)
  return l
}

//Insert : Inserts an element at a particular position specified if not append it add the end of the list
func (l *List) Insert(element int, at int) {
  // set node value
  node := &Node{
    value: element,
  }
  //check for length of list
  // append item if length is zero or position is greater length
  if l.length < 1 || l.length <= at {
    // append item to the list
    l.list = append(l.list, *node)
    // increase length of list
    l.length++
    return
  }

  // check for a zero or negative value(treat as zero)
  if at < 1 {
    l.list = append(append([]Node{}, *node), l.list...)
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
  l.list = append(append(x, *node), y...)
  l.length = len(l.list)
}

// Remove : Remove takes out an item from list
func (l *List) Remove(element int) {
  // loop through list
  for i, item := range l.list {
    // check for an equal item
    if item.value == element {
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
  if at < 1 {
    return
  }

  // set item position as slice index
  at--
  // append split slice at the given index
  l.list = append(l.list[:at], l.list[at+1:]...)
  l.length = len(l.list)
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
  return l.list[at].value
}

// Contains : Returns true if the list contains the element
func (l *List) Contains(element int) bool {
  // loop through list
  for _, item := range l.list {
    // check for an equal item
    if item.value == element {
      return true
    }
  }
  return false
}

// Clear : Clears the list
func (l *List) Clear() {
  l.list = []Node{}
  l.length = len(l.list)
}

// Next : Returns the next element in the list
func (l *List) Next() interface{} {
  // check for length of list
  if l.length < 1 {
    return nil
  }

  // return the next element in the list
  return l.list[0]
}

//Prev : Returns the previous element in the list
func (l *List) Prev() interface{} {
  // check for length of list
  if l.length < 1 {
    return nil
  }

  // return the previous element in the list
  return l.list[l.length-1]
}
