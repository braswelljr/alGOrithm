package core

import (
	"fmt"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/structures/searchtrees"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// setTreeP drives any of them: insert, remove and search, said in words.
type setTreeP struct {
	name string
	t    orderedSetTree
}

func (p *setTreeP) Snapshot() trace.Snapshot { return p.t.Snapshot() }

func (p *setTreeP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "insert" && op.Name != "remove" && op.Name != "search" {
		return "", unknownOp(op, "insert", "remove", "search")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "insert":
		if !p.t.Insert(v) {
			return fmt.Sprintf("%d is already in the %s", v, p.name), nil
		}
		pushed(r, v)
		return fmt.Sprintf("inserted %d (%d values)", v, p.t.Len()), nil
	case "remove":
		if !p.t.Remove(v) {
			return fmt.Sprintf("%d is not in the %s", v, p.name), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d (%d values)", v, p.t.Len()), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.t.Contains(v)), nil
}

type bstP struct{ t *searchtrees.BST[int] }

func (p *bstP) Snapshot() trace.Snapshot { return p.t.Snapshot() }

func (p *bstP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "min", "max":
		get := p.t.Min
		if op.Name == "max" {
			get = p.t.Max
		}
		if v, ok := get(); ok {
			return fmt.Sprintf("%s is %d", op.Name, v), nil
		}
		return "empty", nil
	case "insert", "remove", "search":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		switch op.Name {
		case "insert":
			if !p.t.Insert(v) {
				return fmt.Sprintf("%d is already in the tree", v), nil
			}
			pushed(r, v)
			return fmt.Sprintf("inserted %d (height %d)", v, p.t.Height()), nil
		case "remove":
			if !p.t.Remove(v) {
				return fmt.Sprintf("%d is not in the tree", v), nil
			}
			popped(r, v)
			return fmt.Sprintf("removed %d (height %d)", v, p.t.Height()), nil
		}
		return fmt.Sprintf("contains %d: %v", v, p.t.Contains(v)), nil
	}
	return "", unknownOp(op, "insert", "remove", "search", "min", "max")
}

type avlP struct {
	t   *searchtrees.AVL[int]
	rec *recorderSwitch
}

func (p *avlP) Snapshot() trace.Snapshot { return p.t.Snapshot() }

func (p *avlP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "insert" && op.Name != "remove" && op.Name != "search" {
		return "", unknownOp(op, "insert", "remove", "search")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	p.rec.r = r
	defer func() { p.rec.r = nil }()
	before := p.t.Rotations
	note := func() string {
		if n := p.t.Rotations - before; n > 0 {
			return fmt.Sprintf(" (%d rotation(s), height %d)", n, p.t.Height())
		}
		return fmt.Sprintf(" (height %d)", p.t.Height())
	}
	switch op.Name {
	case "insert":
		if !p.t.Insert(v) {
			return fmt.Sprintf("%d is already in the tree", v), nil
		}
		pushed(r, v)
		return "inserted " + strconv.Itoa(v) + note(), nil
	case "remove":
		if !p.t.Remove(v) {
			return fmt.Sprintf("%d is not in the tree", v), nil
		}
		popped(r, v)
		return "removed " + strconv.Itoa(v) + note(), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.t.Contains(v)), nil
}

type intervalP struct{ t *searchtrees.IntervalTree }

func (p *intervalP) Snapshot() trace.Snapshot { return p.t.Snapshot() }

func (p *intervalP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "insert" && op.Name != "remove" && op.Name != "overlap" {
		return "", unknownOp(op, "insert", "remove", "overlap")
	}
	a, err := opPair(op)
	if err != nil {
		return "", err
	}
	iv := searchtrees.Interval{Lo: a[0], Hi: a[1]}
	switch op.Name {
	case "insert":
		if iv.Lo > iv.Hi {
			return "", inputError(fmt.Sprintf("%v ends before it starts", iv))
		}
		if !p.t.Insert(iv) {
			return fmt.Sprintf("%v is already in the tree", iv), nil
		}
		pushed(r, iv.Lo)
		return fmt.Sprintf("inserted %v (%d intervals)", iv, p.t.Len()), nil
	case "remove":
		if !p.t.Remove(iv) {
			return fmt.Sprintf("%v is not in the tree", iv), nil
		}
		popped(r, iv.Lo)
		return fmt.Sprintf("removed %v", iv), nil
	}
	return fmt.Sprintf("overlapping %v: %v", iv, p.t.Overlapping(iv)), nil
}
