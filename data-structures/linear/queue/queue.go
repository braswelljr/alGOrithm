package queue

type Queue struct {
  items []interface{}
}

// Init initializes a new queue.
func (queue *Queue) Init() *Queue {
  queue.items = make([]interface{}, 0)
  return queue
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
func (queue *Queue) Enqueue(item interface{}) {
  // append item to the slice
  queue.items = append(queue.items, item)
}

// Dequeue removes the first item from the Queue
func (queue *Queue) Dequeue() {
  // remove the first item from the slice
  queue.items = queue.items[1:]
}

// Size returns the length of the array
func (queue *Queue) Size() int {
  return len(queue.items)
}

// Peek returns the first item in the Queue
func (queue *Queue) Peek() interface{} {
  return queue.items[0]
}

// IsEmpty returns true if the Queue is empty
func (queue *Queue) IsEmpty() bool {
  return queue.Size() == 0
}

// Clear removes all items from the Queue
func (queue *Queue) Clear() {
  queue.items = make([]interface{}, 0)
}

// Items returns the items in the Queue
func (queue *Queue) Items() []interface{} {
  return queue.items
}

// Reverse reverses the order of the items in the Queue
func (queue *Queue) Reverse() {
  for i, j := 0, queue.Size()-1; i < j; i, j = i+1, j-1 {
    queue.items[i], queue.items[j] = queue.items[j], queue.items[i]
  }
}
