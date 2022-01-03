package queue

type Queue struct {
  items []interface{}
}

// Init initializes a new queue.
func (q *Queue) Init() *Queue {
  q.items = make([]interface{}, 0)
  return q
}

// New returns a new queue.
func New(elements ...[]interface{}) *Queue {
  if elements == nil || len(elements) < 2 {
    return new(Queue).Init()
  }

  // covert slice  to node
  // Create a new node.
  var element []interface{}

  // Iterate over the list and create a node for each value.
  for _, item := range elements {
    for _, value := range item {
      //append slice to the end of the list
      element = append(element, value)
    }
  }

  // Create a new node.
  return &Queue{items: element}
}

// Enqueue adds new item to the end of the Queue
func (q *Queue) Enqueue(item interface{}) {
  // append item to the slice
  q.items = append(q.items, item)
}

// Dequeue removes the first item from the Queue
func (q *Queue) Dequeue() {
  // remove the first item from the slice
  q.items = q.items[1:]
}

// Size returns the length of the array
func (q *Queue) Size() int {
  return len(q.items)
}

// Peek returns the first item in the Queue
func (q *Queue) Peek() interface{} {
  return q.items[0]
}

// IsEmpty returns true if the Queue is empty
func (q *Queue) IsEmpty() bool {
  return q.Size() == 0
}

// Clear removes all items from the Queue
func (q *Queue) Clear() {
  q.items = make([]interface{}, 0)
}

// Items returns the items in the Queue
func (q *Queue) Items() []interface{} {
  return q.items
}

// Reverse reverses the order of the items in the Queue
func (q *Queue) Reverse() {
  for i, j := 0, q.Size()-1; i < j; i, j = i+1, j-1 {
    q.items[i], q.items[j] = q.items[j], q.items[i]
  }
}
