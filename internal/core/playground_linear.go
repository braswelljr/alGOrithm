package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type ringP struct {
	b *structures.CircularBuffer[int]
}

func (p *ringP) Snapshot() trace.Snapshot { return p.b.Snapshot() }

func (p *ringP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		before := p.b.Overwritten()
		if !p.b.Push(v) {
			return fmt.Sprintf("full: %d refused", v), nil
		}
		pushed(r, v)
		if p.b.Overwritten() > before {
			return fmt.Sprintf("full: %d overwrote the oldest value", v), nil
		}
		return fmt.Sprintf("pushed %d", v), nil
	case "pop":
		v, ok := p.b.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d, the oldest", v), nil
	case "peek":
		if v, ok := p.b.Peek(); ok {
			return fmt.Sprintf("oldest is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "peek")
}

type minStackP struct{ s *structures.MinStack[int] }

func (p *minStackP) Snapshot() trace.Snapshot { return p.s.Snapshot() }

func (p *minStackP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.s.Push(v)
		pushed(r, v)
		m, _ := p.s.Min()
		return fmt.Sprintf("pushed %d (min %d)", v, m), nil
	case "pop":
		v, ok := p.s.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d", v), nil
	case "peek", "min":
		get := p.s.Peek
		if op.Name == "min" {
			get = p.s.Min
		}
		if v, ok := get(); ok {
			return fmt.Sprintf("%s is %d", op.Name, v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "peek", "min")
}

type monoP struct {
	q *structures.MonotonicQueue[int]
}

func (p *monoP) Snapshot() trace.Snapshot { return p.q.Snapshot() }

func (p *monoP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "enqueue", "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.q.Enqueue(v)
		pushed(r, v)
		m, _ := p.q.Max()
		return fmt.Sprintf("enqueued %d (max %d)", v, m), nil
	case "dequeue", "pop":
		v, ok := p.q.Dequeue()
		if !ok {
			return "empty: nothing to dequeue", nil
		}
		popped(r, v)
		return fmt.Sprintf("dequeued %d", v), nil
	case "max":
		if v, ok := p.q.Max(); ok {
			return "max is " + strconv.Itoa(v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "enqueue", "dequeue", "max")
}

type stackP struct{ s *structures.Stack[int] }

func (p *stackP) Snapshot() trace.Snapshot { return p.s.Snapshot() }

func (p *stackP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.s.Push(v)
		pushed(r, v)
		return fmt.Sprintf("pushed %d on top", v), nil
	case "pop":
		v, ok := p.s.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d, the last one in", v), nil
	case "peek":
		if v, ok := p.s.Peek(); ok {
			return fmt.Sprintf("top is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "peek")
}

type queueP struct{ q *structures.Queue[int] }

func (p *queueP) Snapshot() trace.Snapshot { return p.q.Snapshot() }

func (p *queueP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "enqueue", "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.q.Enqueue(v)
		pushed(r, v)
		return fmt.Sprintf("enqueued %d at the back", v), nil
	case "dequeue", "pop":
		v, ok := p.q.Dequeue()
		if !ok {
			return "empty: nothing to dequeue", nil
		}
		popped(r, v)
		return fmt.Sprintf("dequeued %d, the first one in", v), nil
	case "peek":
		if v, ok := p.q.Peek(); ok {
			return fmt.Sprintf("front is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "enqueue", "dequeue", "peek")
}

type dequeP struct{ d *structures.Deque[int] }

func (p *dequeP) Snapshot() trace.Snapshot { return p.d.Snapshot() }

func (p *dequeP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push-front", "push-back":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "push-front" {
			p.d.PushFront(v)
		} else {
			p.d.PushBack(v)
		}
		pushed(r, v)
		return fmt.Sprintf("pushed %d at the %s", v, strings.TrimPrefix(op.Name, "push-")), nil
	case "pop-front", "pop-back":
		pop := p.d.PopFront
		if op.Name == "pop-back" {
			pop = p.d.PopBack
		}
		v, ok := pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d from the %s", v, strings.TrimPrefix(op.Name, "pop-")), nil
	case "front", "back":
		get := p.d.Front
		if op.Name == "back" {
			get = p.d.Back
		}
		if v, ok := get(); ok {
			return fmt.Sprintf("%s is %d", op.Name, v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push-front", "push-back", "pop-front", "pop-back", "front", "back")
}

type listP struct{ l *structures.LinkedList[int] }

func (p *listP) Snapshot() trace.Snapshot { return p.l.Snapshot() }

func (p *listP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name == "reverse" {
		p.l.Reverse()
		return "reversed every link", nil
	}
	if op.Name != "append" && op.Name != "prepend" && op.Name != "remove" && op.Name != "search" && op.Name != "insert" {
		return "", unknownOp(op, "append", "prepend", "remove", "search", "reverse")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "append", "insert":
		p.l.Append(v)
		pushed(r, v)
		return fmt.Sprintf("appended %d at the tail", v), nil
	case "prepend":
		p.l.Prepend(v)
		pushed(r, v)
		return fmt.Sprintf("prepended %d at the head", v), nil
	case "remove":
		if !p.l.Remove(v) {
			return fmt.Sprintf("%d is not in the list", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	}
	if i := p.l.Index(v); i >= 0 {
		return fmt.Sprintf("found %d at position %d after walking %d links", v, i, i), nil
	}
	return fmt.Sprintf("%d is not in the list", v), nil
}

type dlistP struct {
	l *structures.DoublyLinkedList[int]
}

func (p *dlistP) Snapshot() trace.Snapshot { return p.l.Snapshot() }

func (p *dlistP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push-front", "push-back", "remove":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		switch op.Name {
		case "push-front":
			p.l.PushFront(v)
		case "push-back":
			p.l.PushBack(v)
		default:
			if !p.l.Remove(v) {
				return fmt.Sprintf("%d is not in the list", v), nil
			}
			popped(r, v)
			return fmt.Sprintf("removed %d", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("pushed %d at the %s", v, strings.TrimPrefix(op.Name, "push-")), nil
	case "pop-front", "pop-back":
		pop := p.l.PopFront
		if op.Name == "pop-back" {
			pop = p.l.PopBack
		}
		v, ok := pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d", v), nil
	}
	return "", unknownOp(op, "push-front", "push-back", "pop-front", "pop-back", "remove")
}

type sortedP struct{ l *structures.SortedList[int] }

func (p *sortedP) Snapshot() trace.Snapshot { return p.l.Snapshot() }

func (p *sortedP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "insert" && op.Name != "remove" && op.Name != "search" {
		return "", unknownOp(op, "insert", "remove", "search")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "insert":
		i := p.l.Insert(v)
		pushed(r, v)
		return fmt.Sprintf("inserted %d at position %d", v, i), nil
	case "remove":
		if !p.l.Remove(v) {
			return fmt.Sprintf("%d is not there", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	}
	if i := p.l.Index(v); i >= 0 {
		return fmt.Sprintf("found %d at %d by binary search", v, i), nil
	}
	return fmt.Sprintf("%d is not there", v), nil
}
