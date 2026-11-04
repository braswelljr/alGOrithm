package core

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
	"github.com/braswelljr/alGOrithm/pkg/structures/spatial"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type pqP struct{ h *heaps.PriorityHeap[int] }

func (p *pqP) Snapshot() trace.Snapshot { return p.h.Snapshot() }

func (p *pqP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push", "insert":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.h.Push(v)
		pushed(r, v)
		return fmt.Sprintf("pushed %d", v), nil
	case "pop":
		v, ok := p.h.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d, the most urgent", v), nil
	case "top", "peek":
		if v, ok := p.h.Top(); ok {
			return fmt.Sprintf("most urgent is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "top")
}

type heapP struct{ h *heaps.MinHeap[int] }

func (p *heapP) Snapshot() trace.Snapshot { return p.h.Snapshot() }

func (p *heapP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push", "insert":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.h.Push(v)
		pushed(r, v)
		return fmt.Sprintf("pushed %d and sifted it up", v), nil
	case "pop":
		v, ok := p.h.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d, the smallest", v), nil
	case "top", "peek":
		if v, ok := p.h.Top(); ok {
			return fmt.Sprintf("smallest is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "top")
}

type kdP struct{ t *spatial.KDTree }

func (p *kdP) Snapshot() trace.Snapshot { return p.t.Snapshot() }

func (p *kdP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "insert", "nearest":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		pt := spatial.Point{X: a[0], Y: a[1]}
		if op.Name == "insert" {
			if !p.t.Insert(pt) {
				return fmt.Sprintf("%v is already in the tree", pt), nil
			}
			pushed(r, a[0])
			return fmt.Sprintf("inserted %v (%d points)", pt, p.t.Len()), nil
		}
		best, d, visited, ok := p.t.Nearest(pt)
		if !ok {
			return "the tree is empty", nil
		}
		return fmt.Sprintf("nearest to %v is %v at distance² %d, visiting %d of %d nodes", pt, best, d, visited, p.t.Len()), nil
	case "range":
		a, err := opInts(op, 4)
		if err != nil {
			return "", err
		}
		rect := spatial.Rect{Min: spatial.Point{X: min(a[0], a[2]), Y: min(a[1], a[3])}, Max: spatial.Point{X: max(a[0], a[2]), Y: max(a[1], a[3])}}
		return fmt.Sprintf("points in %v: %v", rect, p.t.InRect(rect)), nil
	}
	return "", unknownOp(op, "insert", "nearest", "range")
}

type quadP struct{ q *spatial.QuadTree }

func (p *quadP) Snapshot() trace.Snapshot { return p.q.Snapshot() }

func (p *quadP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "insert":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		pt := spatial.Point{X: a[0], Y: a[1]}
		if !p.q.Insert(pt) {
			return fmt.Sprintf("%v is outside the square or already there", pt), nil
		}
		pushed(r, a[0])
		return fmt.Sprintf("inserted %v (%d points)", pt, p.q.Len()), nil
	case "query":
		a, err := opInts(op, 4)
		if err != nil {
			return "", err
		}
		rect := spatial.Rect{Min: spatial.Point{X: min(a[0], a[2]), Y: min(a[1], a[3])}, Max: spatial.Point{X: max(a[0], a[2]), Y: max(a[1], a[3])}}
		found, cells := p.q.Query(rect)
		return fmt.Sprintf("points in %v: %v, looking in %d cells", rect, found, cells), nil
	}
	return "", unknownOp(op, "insert", "query")
}
