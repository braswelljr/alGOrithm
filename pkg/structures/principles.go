package structures

// The access principles: in what order a container gives values back.
//
//	LIFO  last in, first out   = FILO, first in, last out: a Stack
//	FIFO  first in, first out  = LILO, last in, last out:  a Queue
//	both ends                                              a Deque
//	by priority, smallest first                            a heaps.MinHeap
//
// The interfaces below name each principle, so code can depend on the
// principle rather than on one structure (the dependency-inversion idea): a
// depth-first search needs a LIFO and a breadth-first search a FIFO, whichever
// type provides it. Priority order is heaps.PriorityQueue, beside the heap
// that implements it. The compile-time checks at the bottom keep the structures
// honest.

// LIFO is a last-in, first-out container (also called FILO): Pop returns the
// value pushed most recently.
type LIFO[T any] interface {
	Push(v T)
	Pop() (T, bool)
	Peek() (T, bool)
	Len() int
}

// FIFO is a first-in, first-out container (also called LILO): Dequeue returns
// the value enqueued longest ago.
type FIFO[T any] interface {
	Enqueue(v T)
	Dequeue() (T, bool)
	Peek() (T, bool)
	Len() int
}

// AsStack uses a Deque as a LIFO, pushing and popping at its back.
func AsStack[T any](d *Deque[T]) LIFO[T] { return dequeStack[T]{d} }

// AsQueue uses a Deque as a FIFO, adding at the back and taking from the front.
func AsQueue[T any](d *Deque[T]) FIFO[T] { return dequeQueue[T]{d} }

type dequeStack[T any] struct{ d *Deque[T] }

func (s dequeStack[T]) Push(v T)        { s.d.PushBack(v) }
func (s dequeStack[T]) Pop() (T, bool)  { return s.d.PopBack() }
func (s dequeStack[T]) Peek() (T, bool) { return s.d.Back() }
func (s dequeStack[T]) Len() int        { return s.d.Len() }

type dequeQueue[T any] struct{ d *Deque[T] }

func (q dequeQueue[T]) Enqueue(v T)        { q.d.PushBack(v) }
func (q dequeQueue[T]) Dequeue() (T, bool) { return q.d.PopFront() }
func (q dequeQueue[T]) Peek() (T, bool)    { return q.d.Front() }
func (q dequeQueue[T]) Len() int           { return q.d.Len() }

// Drain empties c in the order its principle gives and returns the values:
// the one function that shows LIFO, FIFO and priority order side by side.
func Drain[T any](pop func() (T, bool)) []T {
	var out []T
	for {
		v, ok := pop()
		if !ok {
			return out
		}
		out = append(out, v)
	}
}

var (
	_ LIFO[int] = (*Stack[int])(nil)
	_ FIFO[int] = (*Queue[int])(nil)
)
