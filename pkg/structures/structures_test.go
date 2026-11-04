package structures_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures"
	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func labels(s trace.Snapshot) []string {
	out := make([]string, len(s.Nodes))
	for i, n := range s.Nodes {
		out[i] = n.Label
	}
	return out
}

func TestStack(t *testing.T) {
	s := structures.NewStack[int]()
	if _, ok := s.Pop(); ok {
		t.Fatal("Pop on empty")
	}
	if _, ok := s.Peek(); ok {
		t.Fatal("Peek on empty")
	}
	s.Push(1)
	s.Push(2)
	if v, _ := s.Peek(); v != 2 {
		t.Errorf("Peek = %d", v)
	}
	if snap := s.Snapshot(); snap.Nodes[1].Role != "top" {
		t.Errorf("snapshot %+v", snap)
	}
	if v, ok := s.Pop(); !ok || v != 2 || s.Len() != 1 {
		t.Errorf("Pop = %d, %v", v, ok)
	}
	if got := structures.NewStack(1, 2, 3).Values(); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("NewStack(1,2,3) = %v", got)
	}
}

func TestQueueAndDeque(t *testing.T) {
	q := structures.NewQueue("a", "b")
	q.Enqueue("c")
	if v, _ := q.Dequeue(); v != "a" {
		t.Errorf("Dequeue = %q", v)
	}
	if snap := q.Snapshot(); !reflect.DeepEqual(labels(snap), []string{"b", "c"}) || snap.Nodes[0].Role != "front" {
		t.Errorf("snapshot %+v", snap)
	}
	d := structures.NewDeque[int]()
	if _, ok := d.PopBack(); ok {
		t.Fatal("PopBack on empty")
	}
	var model []int
	for i := 0; i < 1000; i++ {
		switch i % 5 {
		case 0, 1:
			d.PushBack(i)
			model = append(model, i)
		case 2:
			d.PushFront(i)
			model = append([]int{i}, model...)
		case 3:
			if v, ok := d.PopFront(); ok {
				if v != model[0] {
					t.Fatalf("PopFront = %d, want %d", v, model[0])
				}
				model = model[1:]
			}
		case 4:
			if v, ok := d.PopBack(); ok {
				if v != model[len(model)-1] {
					t.Fatalf("PopBack = %d", v)
				}
				model = model[:len(model)-1]
			}
		}
	}
	if !reflect.DeepEqual(d.Values(), model) {
		t.Error("deque differs from the model")
	}
}

func TestLinkedLists(t *testing.T) {
	l := structures.NewLinkedList(1, 2, 3, 4)
	l.Remove(4) // the tail
	l.Remove(1) // the head
	if tail, _ := l.Tail(); tail != 3 || l.Len() != 2 {
		t.Errorf("after removals: %v, tail %d", l.Values(), tail)
	}
	l.Append(9)
	l.Prepend(0)
	l.Reverse()
	if got := l.Values(); !reflect.DeepEqual(got, []int{9, 3, 2, 0}) {
		t.Errorf("reversed %v", got)
	}
	if h, _ := l.Head(); h != 9 || l.Index(2) != 2 || l.Index(7) != -1 {
		t.Error("Head/Index wrong")
	}
	for _, v := range []int{9, 3, 2, 0} {
		l.Remove(v)
	}
	if _, ok := l.Head(); ok || len(l.Snapshot().Nodes) != 0 {
		t.Error("emptied list not empty")
	}
	d := structures.NewDoublyLinkedList(1, 2, 3)
	d.PushFront(0)
	d.Remove(2)
	if !reflect.DeepEqual(d.Values(), []int{0, 1, 3}) || !reflect.DeepEqual(d.Backward(), []int{3, 1, 0}) {
		t.Errorf("doubly: %v / %v", d.Values(), d.Backward())
	}
	if v, _ := d.PopBack(); v != 3 {
		t.Errorf("PopBack = %d", v)
	}
	if len(d.Snapshot().Edges) != 2 {
		t.Errorf("doubly snapshot edges %d", len(d.Snapshot().Edges))
	}
}

func TestSortedList(t *testing.T) {
	l := structures.NewSortedList(5, 1, 4, 1)
	if !reflect.DeepEqual(l.Values(), []int{1, 1, 4, 5}) {
		t.Fatalf("values %v", l.Values())
	}
	if l.Insert(3) != 2 || l.Index(4) != 3 || l.Index(2) != -1 || !l.Remove(1) || l.Len() != 4 {
		t.Error("Insert/Index/Remove wrong")
	}
}

func TestPrinciples(t *testing.T) {
	in := []int{3, 1, 4, 1, 5}
	stack, queue, heap := structures.NewStack[int](), structures.NewQueue[int](), heaps.NewMinHeap[int]()
	ds, dq := structures.AsStack(structures.NewDeque[int]()), structures.AsQueue(structures.NewDeque[int]())
	for _, v := range in {
		stack.Push(v)
		queue.Enqueue(v)
		heap.Push(v)
		ds.Push(v)
		dq.Enqueue(v)
	}
	lifo := []int{5, 1, 4, 1, 3}
	fifo := []int{3, 1, 4, 1, 5}
	if got := structures.Drain(stack.Pop); !reflect.DeepEqual(got, lifo) {
		t.Errorf("stack (LIFO) = %v", got)
	}
	if got := structures.Drain(ds.Pop); !reflect.DeepEqual(got, lifo) {
		t.Errorf("deque as stack = %v", got)
	}
	if got := structures.Drain(queue.Dequeue); !reflect.DeepEqual(got, fifo) {
		t.Errorf("queue (FIFO) = %v", got)
	}
	if got := structures.Drain(dq.Dequeue); !reflect.DeepEqual(got, fifo) {
		t.Errorf("deque as queue = %v", got)
	}
	if got := structures.Drain(heap.Pop); !reflect.DeepEqual(got, []int{1, 1, 3, 4, 5}) {
		t.Errorf("heap (priority) = %v", got)
	}
}

func TestCircularBuffer(t *testing.T) {
	b := structures.NewCircularBuffer[int](3, false)
	for i := 1; i <= 3; i++ {
		b.Push(i)
	}
	if b.Push(4) || !b.Full() {
		t.Fatal("a full rejecting buffer accepted a value")
	}
	if v, _ := b.Pop(); v != 1 {
		t.Errorf("Pop = %d", v)
	}
	b.Push(4) // wraps into slot 0
	if s := b.Snapshot(); s.Nodes[0].Role != "tail" || s.Nodes[1].Role != "head" || !reflect.DeepEqual(b.Values(), []int{2, 3, 4}) {
		t.Errorf("wrapped buffer: %v %+v", b.Values(), s.Nodes)
	}
	o := structures.NewCircularBuffer[int](2, true)
	for i := 1; i <= 5; i++ {
		o.Push(i)
	}
	if !reflect.DeepEqual(o.Values(), []int{4, 5}) || o.Overwritten() != 3 {
		t.Errorf("overwriting buffer: %v, %d overwritten", o.Values(), o.Overwritten())
	}
	e := structures.NewCircularBuffer[int](2, true)
	if _, ok := e.Pop(); ok {
		t.Error("Pop on empty")
	}
	if s := e.Snapshot(); s.Nodes[0].Role != "free" {
		t.Error("empty slots should be free")
	}
}

func TestMinStackAndMonotonicQueue(t *testing.T) {
	s := structures.NewMinStack(5, 3, 7, 3, 1)
	for _, want := range []int{1, 3, 3, 3, 5} {
		if m, _ := s.Min(); m != want {
			t.Fatalf("Min = %d, want %d", m, want)
		}
		s.Pop()
	}
	if _, ok := s.Min(); ok {
		t.Error("Min on empty")
	}
	in := []int{1, 3, -1, -3, 5, 3, 6, 7}
	q := structures.NewMonotonicQueue[int]()
	var got []int
	for i, v := range in {
		q.Enqueue(v)
		if q.Len() > 3 {
			q.Dequeue()
		}
		if i >= 2 {
			m, _ := q.Max()
			got = append(got, m)
		}
	}
	if !reflect.DeepEqual(got, []int{3, 3, 5, 5, 6, 7}) {
		t.Errorf("window maxima %v", got)
	}
	big := structures.NewMonotonicQueue[int]()
	var model []int
	r := rand.New(rand.NewSource(3))
	for step := 0; step < 2000; step++ {
		if r.Intn(3) > 0 {
			v := r.Intn(100)
			big.Enqueue(v)
			model = append(model, v)
		} else if v, ok := big.Dequeue(); ok {
			if v != model[0] {
				t.Fatal("dequeue order")
			}
			model = model[1:]
		}
		if len(model) > 0 {
			mx := model[0]
			for _, v := range model {
				mx = max(mx, v)
			}
			if m, _ := big.Max(); m != mx {
				t.Fatalf("Max = %d, want %d", m, mx)
			}
			if p, _ := big.Peek(); p != model[0] {
				t.Fatal("Peek")
			}
		}
	}
	if snap := structures.NewMonotonicQueue(1, 3, 2).Snapshot(); snap.Nodes[1].Role != "max" || snap.Nodes[2].Role != "candidate" || snap.Nodes[0].Role != "" {
		t.Errorf("snapshot roles %+v", snap.Nodes)
	}
}

// TestDequeAgainstASlice runs random operations on a Deque, on a LIFO and a
// FIFO view of one, and on a DoublyLinkedList, checking each against a slice.
func TestDequeAgainstASlice(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	d := structures.NewDeque[int]()
	l := structures.NewDoublyLinkedList[int]()
	var model []int
	for step := 0; step < 5000; step++ {
		v := r.Intn(1000)
		switch r.Intn(6) {
		case 0:
			d.PushFront(v)
			l.PushFront(v)
			model = append([]int{v}, model...)
		case 1:
			d.PushBack(v)
			l.PushBack(v)
			model = append(model, v)
		case 2:
			got, ok := d.PopFront()
			lv, lok := l.PopFront()
			if ok != (len(model) > 0) || lok != ok {
				t.Fatalf("step %d: PopFront ok %v %v with %d values", step, ok, lok, len(model))
			}
			if ok {
				if got != model[0] || lv != model[0] {
					t.Fatalf("step %d: PopFront %d %d, want %d", step, got, lv, model[0])
				}
				model = model[1:]
			}
		case 3:
			got, ok := d.PopBack()
			lv, lok := l.PopBack()
			if ok != (len(model) > 0) || lok != ok {
				t.Fatalf("step %d: PopBack ok %v %v", step, ok, lok)
			}
			if ok {
				if got != model[len(model)-1] || lv != got {
					t.Fatalf("step %d: PopBack %d %d, want %d", step, got, lv, model[len(model)-1])
				}
				model = model[:len(model)-1]
			}
		default:
			f, fok := d.Front()
			b, bok := d.Back()
			if fok != (len(model) > 0) || bok != fok {
				t.Fatalf("step %d: Front/Back ok %v %v", step, fok, bok)
			}
			if fok && (f != model[0] || b != model[len(model)-1]) {
				t.Fatalf("step %d: front %d back %d, want %d and %d", step, f, b, model[0], model[len(model)-1])
			}
		}
		if d.Len() != len(model) || l.Len() != len(model) {
			t.Fatalf("step %d: len %d %d, want %d", step, d.Len(), l.Len(), len(model))
		}
	}
	if !reflect.DeepEqual(d.Values(), model) && len(model) > 0 {
		t.Errorf("values differ from the model")
	}
	snap := d.Snapshot()
	if len(snap.Nodes) != len(model) || (len(model) > 1 && (snap.Nodes[0].Role != "front" || snap.Nodes[len(model)-1].Role != "back")) {
		t.Errorf("snapshot roles %v", labels(snap))
	}

	var empty structures.Deque[string]
	if _, ok := empty.Front(); ok {
		t.Error("Front of an empty deque")
	}
	if _, ok := empty.Back(); ok {
		t.Error("Back of an empty deque")
	}
	stack, queue := structures.AsStack(structures.NewDeque[int]()), structures.AsQueue(structures.NewDeque[int]())
	for _, v := range []int{1, 2, 3} {
		stack.Push(v)
		queue.Enqueue(v)
	}
	if top, _ := stack.Peek(); top != 3 || stack.Len() != 3 {
		t.Errorf("stack peek %d len %d", top, stack.Len())
	}
	if front, _ := queue.Peek(); front != 1 || queue.Len() != 3 {
		t.Errorf("queue peek %d len %d", front, queue.Len())
	}
}

func TestSmallAccessors(t *testing.T) {
	q := structures.NewQueue("a", "b")
	if v, ok := q.Peek(); !ok || v != "a" || q.Len() != 2 {
		t.Errorf("queue peek %q len %d", v, q.Len())
	}
	if v, ok := structures.NewQueue[int]().Peek(); ok || v != 0 {
		t.Error("peek of an empty queue")
	}

	b := structures.NewCircularBuffer[int](0, false) // held to 1 slot
	if b.Cap() != 1 || b.Len() != 0 {
		t.Errorf("cap %d len %d", b.Cap(), b.Len())
	}
	if _, ok := b.Peek(); ok {
		t.Error("peek of an empty buffer")
	}
	b.Push(7)
	if v, ok := b.Peek(); !ok || v != 7 || b.Len() != 1 || b.Push(8) {
		t.Errorf("peek %d, len %d, a full non-overwriting buffer took a value", v, b.Len())
	}

	m := structures.NewMinStack(5, 3, 8)
	if v, ok := m.Peek(); !ok || v != 8 || m.Len() != 3 {
		t.Errorf("min stack peek %d len %d", v, m.Len())
	}
	if got := labels(m.Snapshot()); !reflect.DeepEqual(got, []string{"5 (min 5)", "3 (min 3)", "8 (min 3)"}) {
		t.Errorf("min stack snapshot %v", got)
	}
	if m.Snapshot().Nodes[2].Role != "top" {
		t.Error("the top is not marked")
	}
	var em structures.MinStack[int]
	if _, ok := em.Peek(); ok {
		t.Error("peek of an empty min stack")
	}
	if _, ok := em.Pop(); ok {
		t.Error("pop of an empty min stack")
	}

	mq := structures.NewMonotonicQueue[int]()
	if _, ok := mq.Peek(); ok {
		t.Error("peek of an empty monotonic queue")
	}
	if _, ok := mq.Max(); ok {
		t.Error("max of an empty monotonic queue")
	}
	mq.Enqueue(2)
	mq.Enqueue(9)
	if v, _ := mq.Peek(); v != 2 {
		t.Errorf("monotonic peek %d", v)
	}

	ll := structures.NewLinkedList[int]()
	if _, ok := ll.Tail(); ok {
		t.Error("tail of an empty list")
	}
	ll.Prepend(2)
	ll.Prepend(1)
	if tail, _ := ll.Tail(); tail != 2 || !reflect.DeepEqual(ll.Values(), []int{1, 2}) {
		t.Errorf("prepend into an empty list: %v", ll.Values())
	}

	dl := structures.NewDoublyLinkedList(1, 2, 3)
	if dl.Remove(9) || !dl.Remove(3) || !dl.Remove(1) || !reflect.DeepEqual(dl.Backward(), []int{2}) {
		t.Errorf("doubly linked remove: %v", dl.Values())
	}
	dl.PushFront(0)
	if !reflect.DeepEqual(dl.Values(), []int{0, 2}) || dl.Len() != 2 {
		t.Errorf("push front %v", dl.Values())
	}

	sl := structures.NewSortedList(3, 1, 2)
	if sl.Remove(7) {
		t.Error("removed a missing value")
	}
	snap := sl.Snapshot()
	if got := labels(snap); !reflect.DeepEqual(got, []string{"1", "2", "3"}) || snap.Nodes[0].Role != "min" || snap.Nodes[2].Role != "max" || len(snap.Edges) != 2 {
		t.Errorf("sorted list snapshot %v", got)
	}
}
